package overflow

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type FilterFunction func(OverflowTransaction) bool

type BlockResult struct {
	Transactions []OverflowTransaction
	Block        flow.Block
	Error        error
	Logger       *zap.Logger
}

type OverflowTransaction struct {
	Id               flow.Identifier
	Events           OverflowEvents
	Error            error
	Fee              float64
	ExecutionEffort  int
	Status           string
	Arguments        []interface{}
	Stakeholders     map[string][]string
	ProposerKeyIndex int
	Script           []byte
	RawTx            flow.Transaction
}

func (o *OverflowState) GetTransactionResultByBlockId(blockId flow.Identifier) ([]*flow.TransactionResult, error) {
	return o.Services.Transactions.GetTransactionResultsByBlockID(blockId)
}

func (o *OverflowState) GetTransactionByBlockId(blockId flow.Identifier) ([]*flow.Transaction, error) {
	return o.Services.Transactions.GetTransactionsByBlockID(blockId)
}

func (o *OverflowState) GetTransactionById(id flow.Identifier) (*flow.Transaction, error) {
	tx, _, err := o.Services.Transactions.GetStatus(id, false)
	return tx, err
}

func (o *OverflowState) GetTransactions(ctx context.Context, id flow.Identifier) ([]OverflowTransaction, error) {

	//sometimes this will become too complex.

	//if we get this error
	//* rpc error: code = ResourceExhausted desc = grpc: trying to send message larger than max (22072361 vs. 20971520)
	//we have to fetch the block again with transaction ids.
	//in paralell loop over them and run GetStatus and create the transactions that way.

	txR, err := o.GetTransactionResultByBlockId(id)
	if err != nil {
		return nil, errors.Wrap(err, "getting transaction results")
	}

	tx, err := o.GetTransactionByBlockId(id)

	if err != nil {
		return nil, errors.Wrap(err, "getting transactions by id")
	}

	result := lo.FlatMap(txR, func(rp *flow.TransactionResult, i int) []OverflowTransaction {
		r := *rp

		if r.TransactionID.String() == "f31815934bff124e332b3c8be5e1c7a949532707251a9f2f81def8cc9f3d1458" {
			return []OverflowTransaction{}
		}

		t := *tx[i]

		//for some reason we get epoch heartbeat
		if len(t.EnvelopeSignatures) == 0 {
			return []OverflowTransaction{}
		}

		feeAmount := 0.0
		events, fee := parseEvents(r.Events)
		feeRaw, ok := fee.Fields["amount"]
		if ok {
			feeAmount = feeRaw.(float64)
		}

		executionEffort, ok := fee.Fields["executionEffort"].(float64)
		gas := 0
		if ok {
			factor := 100000000
			gas = int(math.Round(executionEffort * float64(factor)))
		}

		args := []interface{}{}
		for i := range t.Arguments {
			arg, err := t.Argument(i)
			if err != nil {
				fmt.Println("[WARN]", err.Error())
			}
			args = append(args, CadenceValueToInterface(arg))
		}

		standardStakeholders := map[string][]string{}
		for _, authorizer := range t.Authorizers {
			standardStakeholders[authorizer.Hex()] = []string{"authorizer"}
		}

		payerRoles, ok := standardStakeholders[t.Payer.Hex()]
		if !ok {
			standardStakeholders[t.Payer.Hex()] = []string{"payer"}
		} else {
			payerRoles = append(payerRoles, "payer")
			standardStakeholders[t.Payer.Hex()] = payerRoles
		}

		proposer, ok := standardStakeholders[t.ProposalKey.Address.Hex()]
		if !ok {
			standardStakeholders[t.ProposalKey.Address.Hex()] = []string{"proposer"}
		} else {
			proposer = append(proposer, "proposer")
			standardStakeholders[t.ProposalKey.Address.Hex()] = proposer
		}

		eventsWithoutFees := events.FilterFees(feeAmount)
		return []OverflowTransaction{{
			Id:               r.TransactionID,
			Status:           r.Status.String(),
			Events:           eventsWithoutFees,
			Stakeholders:     eventsWithoutFees.GetStakeholders(standardStakeholders),
			Error:            r.Error,
			Arguments:        args,
			Fee:              feeAmount,
			Script:           t.Script,
			ProposerKeyIndex: t.ProposalKey.KeyIndex,
			ExecutionEffort:  gas,
			RawTx:            t,
		}}
	})

	return result, nil

}

// This code is beta
func (o *OverflowState) StreamTransactions(ctx context.Context, poll time.Duration, height uint64, logger *zap.Logger, channel chan<- BlockResult) error {

	latestKnownBlock, err := o.GetLatestBlock()
	if err != nil {
		return err
	}

	sleep := poll
	for {
		select {
		case <-time.After(sleep):

			sleep = poll
			nextBlockToProcess := height + 1
			if height == uint64(0) {
				nextBlockToProcess = latestKnownBlock.Height
				height = latestKnownBlock.Height
			}
			logg := logger.With(zap.Uint64("height", height), zap.Uint64("nextBlockToProcess", nextBlockToProcess), zap.Uint64("latestKnownBlock", latestKnownBlock.Height))

			var block *flow.Block
			if nextBlockToProcess < latestKnownBlock.Height {
				//we are still processing historical blocks
				block, err = o.GetBlockAtHeight(nextBlockToProcess)
				if err != nil {
					logg.Debug("error fetching old block", zap.Error(err))
					continue
				}
			} else if nextBlockToProcess != latestKnownBlock.Height {
				block, err = o.GetLatestBlock()
				if err != nil {
					logg.Debug("error fetching latest block", zap.Error(err))
					continue
				}

				if block == nil || block.Height == latestKnownBlock.Height {
					continue
				}
				latestKnownBlock = block
				//we just continue the next iteration in the loop here
				sleep = time.Millisecond
				//the reason we just cannot process here is that the latestblock might not be the next block we should process
				continue
			} else {
				block = latestKnownBlock
			}
			tx, err := o.GetTransactions(ctx, block.ID)
			if err != nil {
				logg.Debug("getting transaction", zap.Error(err))
				if strings.Contains(err.Error(), "could not retrieve collection: key not found") {
					continue
				}
				channel <- BlockResult{Block: *block, Error: errors.Wrap(err, "getting transactions"), Logger: logg}
				height = nextBlockToProcess
				continue
			}
			logg = logg.With(zap.Int("tx", len(tx)))

			channel <- BlockResult{Block: *block, Transactions: tx, Logger: logg}
			height = nextBlockToProcess

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

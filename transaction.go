package overflow

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/cadence/runtime/parser"
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

type Argument struct {
	Key   string
	Value interface{}
}

type OverflowTransaction struct {
	Id              flow.Identifier
	BlockId         flow.Identifier
	Events          []OverflowEvent
	Error           error
	Fee             float64
	Status          string
	Arguments       []Argument
	Authorizers     []string
	Stakeholders    map[string][]string
	Imports         []Import
	Payer           string
	ProposalKey     flow.ProposalKey
	GasLimit        uint64
	GasUsed         uint64
	ExecutionEffort float64
	Script          []byte
}

func (o *OverflowState) GetTransactionById(ctx context.Context, id flow.Identifier) (*flow.Transaction, error) {
	tx, _, err := o.Flowkit.GetTransactionByID(ctx, id, false)
	return tx, err
}

func (o *OverflowState) GetTransactions(ctx context.Context, id flow.Identifier) ([]OverflowTransaction, error) {

	//sometimes this will become too complex.

	//if we get this error
	//* rpc error: code = ResourceExhausted desc = grpc: trying to send message larger than max (22072361 vs. 20971520)
	//we have to fetch the block again with transaction ids.
	//in parallel loop over them and run GetStatus and create the transactions that way.

	tx, txR, err := o.Flowkit.GetTransactionsByBlockID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "getting transaction results")
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
			feeAmount, ok = feeRaw.(float64)
			if !ok {
				panic("failed casting fee amount to float64")
			}
		}

		executionEffort, ok := fee.Fields["executionEffort"].(float64)
		gas := 0
		if ok {
			factor := 100000000
			gas = int(math.Round(executionEffort * float64(factor)))
		}

		args := []Argument{}
		argInfo := declarationInfo(t.Script)
		for i := range t.Arguments {
			arg, err := t.Argument(i)
			if err != nil {
				fmt.Println("[WARN]", err.Error())
			}
			argStruct := Argument{
				Key:   argInfo.ParameterOrder[i],
				Value: CadenceValueToInterface(arg),
			}
			args = append(args, argStruct)
		}

		standardStakeholders := map[string][]string{}
		imports, err := GetAddressImports(t.Script)
		if err != nil {
			fmt.Println("[WARN]", err.Error())
		}

		authorizers := []string{}
		for _, authorizer := range t.Authorizers {
			auth := fmt.Sprintf("0x%s", authorizer.Hex())
			authorizers = append(authorizers, auth)
			standardStakeholders[auth] = []string{"authorizer"}
		}

		payerRoles, ok := standardStakeholders[fmt.Sprintf("0x%s", t.Payer.Hex())]
		if !ok {
			standardStakeholders[fmt.Sprintf("0x%s", t.Payer.Hex())] = []string{"payer"}
		} else {
			payerRoles = append(payerRoles, "payer")
			standardStakeholders[fmt.Sprintf("0x%s", t.Payer.Hex())] = payerRoles
		}

		proposer, ok := standardStakeholders[fmt.Sprintf("0x%s", t.ProposalKey.Address.Hex())]
		if !ok {
			standardStakeholders[fmt.Sprintf("0x%s", t.ProposalKey.Address.Hex())] = []string{"proposer"}
		} else {
			proposer = append(proposer, "proposer")
			standardStakeholders[fmt.Sprintf("0x%s", t.ProposalKey.Address.Hex())] = proposer
		}

		eventsWithoutFees := events.FilterFees(feeAmount)

		eventList := []OverflowEvent{}
		for _, evList := range eventsWithoutFees {
			eventList = append(eventList, evList...)
		}

		return []OverflowTransaction{{
			Id:              r.TransactionID,
			BlockId:         id,
			Status:          r.Status.String(),
			Events:          eventList,
			Stakeholders:    eventsWithoutFees.GetStakeholders(standardStakeholders),
			Imports:         imports,
			Error:           r.Error,
			Arguments:       args,
			Fee:             feeAmount,
			Script:          t.Script,
			Payer:           fmt.Sprintf("0x%s", t.Payer.String()),
			ProposalKey:     t.ProposalKey,
			GasLimit:        t.GasLimit,
			GasUsed:         uint64(gas),
			ExecutionEffort: executionEffort,
			Authorizers:     authorizers,
		}}
	})

	return result, nil

}

// This code is beta
func (o *OverflowState) StreamTransactions(ctx context.Context, poll time.Duration, height uint64, logger *zap.Logger, channel chan<- BlockResult) error {

	latestKnownBlock, err := o.GetLatestBlock(ctx)
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
				block, err = o.GetBlockAtHeight(ctx, nextBlockToProcess)
				if err != nil {
					logg.Debug("error fetching old block", zap.Error(err))
					continue
				}
			} else if nextBlockToProcess != latestKnownBlock.Height {
				block, err = o.GetLatestBlock(ctx)
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

func GetAddressImports(code []byte) ([]Import, error) {

	deps := []Import{}
	program, err := parser.ParseProgram(nil, code, parser.Config{})
	if err != nil {
		return deps, err
	}

	for _, imp := range program.ImportDeclarations() {
		address, isAddressImport := imp.Location.(common.AddressLocation)
		if isAddressImport {
			for _, id := range imp.Identifiers {

				deps = append(deps, Import{
					Address: fmt.Sprintf("0x%s", address.Address.Hex()),
					Name:    id.Identifier,
				})
			}
		}
	}
	return deps, nil
}

type Import struct {
	Address string
	Name    string
}

func (i Import) Identifier() string {
	return fmt.Sprintf("A.%s.%s", strings.TrimPrefix(i.Address, "0x"), i.Name)

}

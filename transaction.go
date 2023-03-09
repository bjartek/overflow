package overflow

import (
	"context"
	"fmt"
	"math"

	"github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type FilterFunction func(OverflowTransaction) bool

type BlockResult struct {
	Transactions []OverflowTransaction
	Block        *flow.Block
	Error        error
}

type OverflowTransaction struct {
	Id              flow.Identifier
	Events          OverflowEvents
	Error           error
	Fee             float64
	ExecutionEffort int
	Status          string
	Arguments       []interface{}
	RawTx           flow.Transaction
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
		return []OverflowTransaction{{
			Id:              r.TransactionID,
			Status:          r.Status.String(),
			Events:          events.FilterFees(feeAmount),
			Error:           r.Error,
			Arguments:       args,
			Fee:             feeAmount,
			ExecutionEffort: gas,
			RawTx:           t,
		}}
	})

	return result, nil

}

/**

o := overflow.Overflow(overflow.WithNetwork("mainnet"), overflow.WithPrintResults())
	if o.Error != nil {
		panic(o.Error)
	}

	ctx := context.Background()

	height := uint64(45219437)

	overflowChannel := make(chan overflow.Transaction)

	defer close(overflowChannel)

	go func() {
		o.StreamTransactions(ctx, height, overflowChannel)
	}()

	for res := range overflowChannel {
		fmt.Println(lo.Keys(res.Events))
	}

**/
/*


- filter : joyride
- classify: Deposit events, one transaction can have multiple items
- bulk transform a given classification: transform all Deposit that have Views into NFTDIct
- send to stream
*/
/*
func (o *OverflowState) StreamTransactions(ctx context.Context, height uint64, channel chan<- BlockResult) {

	for {
		block, err := o.GetBlockAtHeight(height)
		if err != nil {
			channel <- BlockResult{Block: block, Error: errors.Wrap(err, "getting block")}
			time.Sleep(1 * time.Second)
			continue
		}
		tx, err := o.GetTransactions(ctx, block.ID)
		if err != nil {
			channel <- BlockResult{Block: block, Error: errors.Wrap(err, "getting transactions")}
			time.Sleep(200 * time.Millisecond)
			continue
		}

		channel <- BlockResult{Block: block, Transactions: tx}
		log.Printf("getting transactions for block id %s height:%d tx:%d\n", block.ID.String(), block.Height, len(tx))

		height = height + 1
	}
}
*/

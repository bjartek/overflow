package overflow

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

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

func (o *OverflowState) GetTransactions(ctx context.Context, id flow.Identifier) ([]OverflowTransaction, error) {

	tx, err := o.GetTransactionByBlockId(id)
	if err != nil {
		return nil, errors.Wrap(err, "getting transactions by id")
	}

	txMap := lo.Associate(tx, func(t *flow.Transaction) (flow.Identifier, flow.Transaction) {
		return t.ID(), *t
	})

	txR, err := o.GetTransactionResultByBlockId(id)
	if err != nil {
		return nil, errors.Wrap(err, "getting transaction results")
	}

	return lo.FlatMap(txR, func(r *flow.TransactionResult, _ int) []OverflowTransaction {
		t := txMap[r.TransactionID]
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
			Status:          r.Status.String(),
			Events:          events.FilterFees(feeAmount),
			Error:           r.Error,
			Arguments:       args,
			Fee:             feeAmount,
			ExecutionEffort: gas,
			RawTx:           t,
		}}
	}), nil

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
func (o *OverflowState) StreamTransactions(ctx context.Context, height uint64, channel chan<- BlockResult) {

	for {
		block, err := o.GetBlockAtHeight(height)
		if err != nil {
			channel <- BlockResult{Block: block, Error: err}
			time.Sleep(1 * time.Second)
			continue
		}
		tx, err := o.GetTransactions(ctx, block.ID)
		if err != nil {
			channel <- BlockResult{Block: block, Error: err}
			time.Sleep(1 * time.Second)
			continue
		}

		channel <- BlockResult{Block: block, Transactions: tx}
		log.Printf("getting transactions for block id %s height:%d tx:%d\n", block.ID.String(), block.Height, len(tx))

		height = height + 1
	}
}

package overflow

import (
	"context"
	"log"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type FilterFunction func(Transaction) bool

type BlockResult struct {
	Transactions []Transaction
	Block        uint64
	Error        error
}

type Transaction struct {
	Events      OverflowEvents
	Fee         OverflowEvent
	RawTxResult flow.TransactionResult
	RawTx       flow.Transaction
}

func (o *OverflowState) GetTransactionResultByBlockId(blockId flow.Identifier) ([]*flow.TransactionResult, error) {
	return o.Services.Transactions.GetTransactionResultsByBlockID(blockId)
}

func (o *OverflowState) GetTransactionByBlockId(blockId flow.Identifier) ([]*flow.Transaction, error) {
	return o.Services.Transactions.GetTransactionsByBlockID(blockId)
}

func (o *OverflowState) GetTransactions(ctx context.Context, id flow.Identifier) ([]Transaction, error) {

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

	return lo.FlatMap(txR, func(r *flow.TransactionResult, _ int) []Transaction {
		t := txMap[r.TransactionID]
		//for some reason we get epoch heartbeat
		if len(t.EnvelopeSignatures) == 0 {
			return []Transaction{}
		}

		feeAmount := 0.0
		events, fee := parseEvents(r.Events)
		feeRaw, ok := fee.Fields["amount"]
		if ok {
			feeAmount = feeRaw.(float64)
		}
		return []Transaction{{
			Events:      events.FilterFees(feeAmount),
			Fee:         fee,
			RawTxResult: *r,
			RawTx:       txMap[r.TransactionID],
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
			channel <- BlockResult{Block: height, Error: err}
			time.Sleep(1 * time.Second)
			continue
		}
		tx, err := o.GetTransactions(ctx, block.ID)
		if err != nil {
			channel <- BlockResult{Block: height, Error: err}
			time.Sleep(1 * time.Second)
			continue
		}

		channel <- BlockResult{Block: height, Transactions: tx}
		log.Printf("getting transactions for block id %s height:%d tx:%d\n", block.ID.String(), block.Height, len(tx))

		height = height + 1
	}
}

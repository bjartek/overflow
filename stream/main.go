package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/bjartek/overflow"
	"go.uber.org/zap"
)

func main() {

	o := overflow.Overflow(overflow.WithNetwork("mainnet"))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	overflowChannel := make(chan overflow.BlockResult)
	defer close(overflowChannel)

	go func() {
		o.StreamTransactions(ctx, time.Second*5, 0, logger, overflowChannel)
	}()

	for {
		select {
		case <-ctx.Done():
			stop()
			break
		case br := <-overflowChannel:
			l := br.Logger
			l.Debug("got stuff")
			l.Sync()
		}
	}
}

package overflow

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/onflow/flow-go/utils/io"
	"github.com/stretchr/testify/require"
)

type OverflowTest struct {
	O      *OverflowState
	height uint64
}

func (ot *OverflowTest) Reset() error {
	block, err := ot.O.GetLatestBlock(context.Background())
	if err != nil {
		return err
	}
	height := block.Height
	if ot.height != height {
		return ot.O.RollbackToBlockHeight(ot.height)
	}
	return nil
}

func (ot *OverflowTest) Run(t *testing.T, name string, f func(t *testing.T)) {
	t.Helper()
	err := ot.Reset()
	require.NoError(t, err)
	t.Run(name, f)
	err = ot.Reset()
	require.NoError(t, err)
}

func (ot *OverflowTest) Teardown() {

	report := ot.O.GetCoverageReport()
	if report == nil {
		return
	}

	bytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		panic(err)
	}
	err = io.WriteFile("coverage-report.json", bytes)
	if err != nil {
		panic(err)
	}

}

func SetupTest(opts []OverflowOption, setup func(o *OverflowState) error) (*OverflowTest, error) {
	allOpts := []OverflowOption{WithNetwork("testing")}
	allOpts = append(allOpts, opts...)

	o := Overflow(allOpts...)
	if o.Error != nil {
		return nil, o.Error
	}

	err := setup(o)
	if err != nil {
		return nil, err
	}

	if o.Error != nil {
		return nil, err
	}

	block, err := o.GetLatestBlock(context.Background())
	if err != nil {
		return nil, err
	}
	height := block.Height

	ot := &OverflowTest{O: o, height: height}
	return ot, nil
}

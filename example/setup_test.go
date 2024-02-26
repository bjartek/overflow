package example

import (
	"os"
	"testing"

	"github.com/bjartek/overflow/v2"
)

// we set the shared overflow test struct that will reset to known setup state after each test
var ot *overflow.OverflowTest

func TestMain(m *testing.M) {
	var err error
	ot, err = overflow.SetupTest([]overflow.OverflowOption{overflow.WithCoverageReport()}, func(o *overflow.OverflowState) error {
		o.MintFlowTokens("first", 1000.0)
		return nil
	})
	if err != nil {
		panic(err)
	}
	code := m.Run()
	ot.Teardown()
	os.Exit(code)
}

package example

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
	// in order to run a test that will reset to known state in setup_test use the `ot.Run(t,...)``method instead of `t.Run(...)`
	ot.Run(t, "Example test", func(t *testing.T) {
		block, err := ot.O.GetLatestBlock(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 4, int(block.Height))

		ot.O.MintFlowTokens("first", 1000.0)
		require.NoError(t, ot.O.Error)

		block, err = ot.O.GetLatestBlock(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 5, int(block.Height))
	})

	ot.Run(t, "Example test 2", func(t *testing.T) {
		block, err := ot.O.GetLatestBlock(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 4, int(block.Height))
	})
}

package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRollback(t *testing.T) {

	o, err := OverflowTesting(WithCoverageReport())
	require.NoError(t, err)
	require.NotNil(t, o)

	block, err := o.GetLatestBlock()
	require.NoError(t, err)
	assert.Equal(t, uint64(7), block.Height)
	o.Tx("mint_tokens", WithSignerServiceAccount(), WithArg("recipient", "first"), WithArg("amount", 1.0)).AssertSuccess(t)

	block, err = o.GetLatestBlock()
	require.NoError(t, err)

	require.NoError(t, err)
	assert.Equal(t, uint64(8), block.Height)

	err = o.RollbackToBlockHeight(6)
	require.NoError(t, err)

	block, err = o.GetLatestBlock()
	require.NoError(t, err)
	assert.Equal(t, uint64(7), block.Height)

}

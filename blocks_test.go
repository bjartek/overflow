package overflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetBlock(t *testing.T) {

	t.Run("Should get latest block", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NoError(t, err)
		require.NotNil(t, g)
		block, err := g.GetLatestBlock(context.Background())

		assert.Nil(t, err)
		assert.GreaterOrEqual(t, block.Height, uint64(0))
	})

	t.Run("Should get block by height", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NoError(t, err)
		block, err := g.GetBlockAtHeight(context.Background(), 0)

		assert.Nil(t, err)
		assert.Equal(t, uint64(0), block.Height)
	})

	t.Run("Should get block by ID", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NoError(t, err)
		block, err := g.GetBlockAtHeight(context.Background(), 0)
		assert.Nil(t, err)
		block, err = g.GetBlockById(context.Background(), block.ID.String())
		assert.Nil(t, err)
		assert.NotNil(t, block)
	})

}

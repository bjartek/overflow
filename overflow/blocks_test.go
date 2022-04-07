package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBlock(t *testing.T) {

	t.Run("Should get latest block", func(t *testing.T) {
		g, _ := NewTestingEmulator().StartE()
		block, err := g.GetLatestBlock()

		assert.Nil(t, err)
		assert.Equal(t, uint64(0), block.Height)
	})

	t.Run("Should get block by height", func(t *testing.T) {
		g, _ := NewTestingEmulator().StartE()
		block, err := g.GetBlockAtHeight(0)

		assert.Nil(t, err)
		assert.Equal(t, uint64(0), block.Height)
	})

	t.Run("Should get block by ID", func(t *testing.T) {
		BlockZeroID := "7bc42fe85d32ca513769a74f97f7e1a7bad6c9407f0d934c2aa645ef9cf613c7"
		g, _ := NewTestingEmulator().StartE()
		block, err := g.GetBlockById(BlockZeroID)

		assert.Nil(t, err)
		assert.Equal(t, BlockZeroID, block.ID.String())
	})

}

package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBlock(t *testing.T) {

	t.Run("Should get latest block", func(t *testing.T) {
		g, _ := OverflowTesting()
		block, err := g.GetLatestBlock()

		assert.Nil(t, err)
		assert.GreaterOrEqual(t, block.Height, uint64(0))
	})

	t.Run("Should get block by height", func(t *testing.T) {
		g, _ := OverflowTesting()
		block, err := g.GetBlockAtHeight(0)

		assert.Nil(t, err)
		assert.Equal(t, uint64(0), block.Height)
	})

	t.Run("Should get block by ID", func(t *testing.T) {
		BlockZeroID := "13c7ff23bb65feb5757cc65fdd75cd243506518c126385fae530ddebdad10b17"
		g, _ := OverflowTesting()
		block, err := g.GetBlockById(BlockZeroID)

		assert.Nil(t, err)
		assert.Equal(t, BlockZeroID, block.ID.String())
	})

}

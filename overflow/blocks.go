package overflow

import (
	"fmt"

	"github.com/onflow/flow-go-sdk"
)

func (f *Overflow) GetLatestBlock() (*flow.Block, error) {
	block, _, _, err := f.Services.Blocks.GetBlock("latest", "", false)
	return block, err
}

func (f *Overflow) GetBlockAtHeight(height uint64) (*flow.Block, error) {
	block, _, _, err := f.Services.Blocks.GetBlock(fmt.Sprintf("%d", height), "", false)
	return block, err
}

// blockId should be a hexadecimal string
func (f *Overflow) GetBlockById(blockId string) (*flow.Block, error) {
	block, _, _, err := f.Services.Blocks.GetBlock(blockId, "", false)
	return block, err
}

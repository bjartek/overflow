package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicIfStopOnFailure(t *testing.T) {
	o, err := OverflowTesting(StopOnError())
	assert.NoError(t, err)

	t.Run("transaction", func(t *testing.T) {
		assert.PanicsWithError(t, "💩 You need to set the main signer", func() {
			o.Tx("create_nft_collection")
		})
	})

	t.Run("script", func(t *testing.T) {

		assert.PanicsWithError(t, "💩 Could not read interaction file from path=./scripts/asdf.cdc", func() {
			o.Script("asdf")
		})
	})
}

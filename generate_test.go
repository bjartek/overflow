package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
 Tests must be in the same folder as flow.json with contracts and transactions/scripts in subdirectories in order for the path resolver to work correctly
*/
func TestGenerate(t *testing.T) {

	o := NewTestingEmulator().Start()
	t.Run("script", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "scripts/test.cdc")
		assert.NoError(t, err)
		assert.Equal(t, "o.Script(\"test\",\n  WithArg(\"account\", \"Address\"),\n)", stub)
	})
	t.Run("script with no arg", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "scripts/type.cdc")
		assert.NoError(t, err)
		assert.Equal(t, "o.Script(\"type\")", stub)
	})

	t.Run("transaction", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "transactions/arguments.cdc")
		assert.NoError(t, err)
		assert.Equal(t, "o.Tx(\"arguments\",\nWithSigner(\"\"),\n  WithArg(\"test\", \"String\"),\n)", stub)
	})

	t.Run("transaction with no args", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "transactions/create_nft_collection.cdc")
		assert.NoError(t, err)
		assert.Equal(t, "o.Tx(\"create_nft_collection\",\nWithSigner(\"\"),\n)", stub)
	})

}

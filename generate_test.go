package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Tests must be in the same folder as flow.json with contracts and transactions/scripts in subdirectories in order for the path resolver to work correctly
*/
func TestGenerate(t *testing.T) {

	o, err := NewTestingEmulator().StartE()
	require.NoError(t, err)
	t.Run("script", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "scripts/test.cdc", false)
		assert.NoError(t, err)
		assert.Equal(t, "  o.Script(\"test\",\n    WithArg(\"account\", <>), //Address\n  )", stub)
	})
	t.Run("script with no arg", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "scripts/type.cdc", false)
		assert.NoError(t, err)
		assert.Equal(t, "  o.Script(\"type\")", stub)
	})

	t.Run("transaction", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "transactions/arguments.cdc", false)
		assert.NoError(t, err)
		assert.Equal(t, "  o.Tx(\"arguments\",\n    WithSigner(\"<>\"),\n    WithArg(\"test\", <>), //String\n  )", stub)
	})

	t.Run("transaction with no args", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "transactions/create_nft_collection.cdc", false)
		assert.NoError(t, err)
		assert.Equal(t, "  o.Tx(\"create_nft_collection\",\n    WithSigner(\"<>\"),\n  )", stub)
	})

	t.Run("transaction standalone", func(t *testing.T) {
		stub, err := o.GenerateStub("emulator", "transactions/arguments.cdc", true)
		assert.NoError(t, err)
		assert.Equal(t, `package main

import (
   . "github.com/bjartek/overflow"
)

func main() {
  o := Overflow(WithNetwork("emulator"), WithPrintResults())
  o.Tx("arguments",
    WithSigner("<>"),
    WithArg("test", <>), //String
  )
}`, stub)

	})

}

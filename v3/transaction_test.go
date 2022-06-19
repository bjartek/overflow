package v3

import (
	"testing"

	"github.com/bjartek/overflow/overflow"
	"github.com/stretchr/testify/assert"
)

func TestTransaction(t *testing.T) {
	o, _ := OverflowTesting()

	t.Run("Run simple tx", func(t *testing.T) {
		res := o.Tx("arguments", Arg("test", "foo"), SignProposeAndPayAsServiceAccount())
		assert.NoError(t, res.Err)
	})

	t.Run("Run linine tx", func(t *testing.T) {
		res := o.Tx(`
transaction(test:String) {
  prepare(acct: AuthAccount) {
    log(test)
 }
}
`, Arg("test", "foo"), SignProposeAndPayAsServiceAccount())
		assert.NoError(t, res.Err)
	})

	t.Run("Run simple tx with custom signer", func(t *testing.T) {
		res := o.Tx("arguments", Arg("test", "foo"), SignProposeAndPayAs("account"))
		assert.NoError(t, res.Err)
	})

	t.Run("Error on wrong signer name", func(t *testing.T) {
		res := o.Tx("arguments", Arg("test", "foo"), SignProposeAndPayAs("account2"))
		assert.ErrorContains(t, res.Err, "could not find account with name emulator-account2")
	})

	t.Run("compose a function", func(t *testing.T) {
		serviceAccountTx := o.TxFN(SignProposeAndPayAsServiceAccount())
		res := serviceAccountTx("arguments", Arg("test", "foo"))
		assert.NoError(t, res.Err)
	})

	t.Run("create function with name", func(t *testing.T) {
		argumentTx := o.TxFileNameFN("arguments", SignProposeAndPayAsServiceAccount())
		res := argumentTx(Arg("test", "foo"))
		assert.NoError(t, res.Err)
	})

	t.Run("Should not allow varags builder arg with single element", func(t *testing.T) {
		res := o.Tx("arguments", Args("test"))
		assert.ErrorContains(t, res.Err, "Please send in an even number of string : interface{} pairs")
	})

	t.Run("Should not allow varag with non string keys", func(t *testing.T) {
		res := o.Tx("arguments", Args(1, "test"))
		assert.ErrorContains(t, res.Err, "even parameters in Args needs to be string")
	})

	t.Run("Arg, with cadence raw value", func(t *testing.T) {
		res := o.Tx("arguments", SignProposeAndPayAsServiceAccount(), CArg("test", overflow.CadenceString("test")))
		assert.NoError(t, res.Err)
	})

	t.Run("Map args", func(t *testing.T) {
		res := o.Tx("arguments", SignProposeAndPayAsServiceAccount(), ArgsM(map[string]interface{}{"test": "test"}))
		assert.NoError(t, res.Err)
	})

}

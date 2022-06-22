package v3

import (
	"fmt"
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

	t.Run("Run simple tx with sa proposer", func(t *testing.T) {
		res := o.Tx("arguments", Arg("test", "foo"), PayloadSigner("first"), ProposeAsServiceAccount())
		assert.Contains(t, res.EmulatorLog[4], "0x01cf0e2f2f715450")
	})

	t.Run("Run simple tx with custom proposer", func(t *testing.T) {
		res := o.Tx("arguments", Arg("test", "foo"), PayloadSigner("first"), ProposeAs("account"))
		assert.Contains(t, res.EmulatorLog[4], "0x01cf0e2f2f715450")
	})

	t.Run("Fail when invalid proposer", func(t *testing.T) {
		res := o.Tx("arguments", Arg("test", "foo"), PayloadSigner("first"), ProposeAs("account2"))
		assert.ErrorContains(t, res.Err, "could not find account with name emulator-account2 in the configuration")
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

	t.Run("Run linine tx", func(t *testing.T) {
		res := o.Tx(`
transaction(test:UInt64) {
  prepare(acct: AuthAccount) {
    log(test)
 }
}
`, Arg("test", 1), SignProposeAndPayAsServiceAccount())
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
		res := o.Tx("arguments", SignProposeAndPayAsServiceAccount(), Arg("test", overflow.CadenceString("test")))
		assert.NoError(t, res.Err)
	})

	t.Run("Map args", func(t *testing.T) {
		res := o.Tx("arguments", SignProposeAndPayAsServiceAccount(), ArgsM(map[string]interface{}{"test": "test"}))
		assert.NoError(t, res.Err)
	})

	t.Run("Parse addresses should fail if not valid account name and hex", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:[Address]) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", Addresses("test", "bjartek"), SignProposeAndPayAsServiceAccount())
		assert.ErrorContains(t, res.Error, "bjartek is not an valid account name or an address")
	})

	t.Run("Parse array of addresses", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:[Address]) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", Addresses("test", "account", "45a1763c93006ca"), SignProposeAndPayAsServiceAccount())
		assert.Equal(t, "[0xf8d6e0586b0a20c7, 0x045a1763c93006ca]", fmt.Sprintf("%v", res.NamedArgs["test"]))
	})

	t.Run("Parse String to String map", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:{String:String}) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", Arg("test", `{ "foo" : "bar"}`), SignProposeAndPayAsServiceAccount())
		assert.Equal(t, `{ "foo" : "bar"}`, fmt.Sprintf("%v", res.NamedArgs["test"]))
	})

	t.Run("Parse String to UFix64 map", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:{String:UFix64}) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", Arg("test", `{ "foo" : 1.0}`), SignProposeAndPayAsServiceAccount())
		assert.Equal(t, `{ "foo" : 1.0}`, fmt.Sprintf("%v", res.NamedArgs["test"]))
	})

	t.Run("Error when parsing invalid address", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:Address) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", Arg("test", "bjartek"), SignProposeAndPayAsServiceAccount())
		assert.ErrorContains(t, res.Error, "argument `test` with value `0xbjartek` is not expected type `Address`")

	})

	t.Run("Should set gas", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:Address) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", Arg("test", "bjartek"), SignProposeAndPayAsServiceAccount(), Gas(100))

		assert.Equal(t, uint64(100), res.GasLimit)

	})

	t.Run("Should report error if invalid payload signer", func(t *testing.T) {
		res := o.Tx(`
transaction{
	prepare(acct: AuthAccount, user:AuthAccount) {

 }
}
`, SignProposeAndPayAsServiceAccount(), PayloadSigner("bjartek"))

		assert.Error(t, res.Err, "asd")

	})
}

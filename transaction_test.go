package overflow

import (
	"fmt"
	"testing"

	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransaction(t *testing.T) {
	o, err := OverflowTesting()
	require.NotNil(t, o)
	require.NoError(t, err)
	t.Run("Run simple tx", func(t *testing.T) {
		res := o.Tx("arguments", WithArg("test", "foo"), WithSignerServiceAccount())
		assert.NoError(t, res.Err)
	})

	t.Run("error on missing argument", func(t *testing.T) {
		res := o.Tx("arguments", WithSignerServiceAccount())
		assert.ErrorContains(t, res.Err, "the interaction 'arguments' is missing [test]")
	})

	t.Run("error on redundant argument", func(t *testing.T) {
		res := o.Tx("arguments", WithArg("test2", "foo"), WithArg("test", "foo"), WithSignerServiceAccount())
		assert.ErrorContains(t, res.Err, "the interaction 'arguments' has the following extra arguments [test2]")
	})

	t.Run("Run simple tx with sa proposer", func(t *testing.T) {
		res := o.Tx("arguments", WithArg("test", "foo"), WithPayloadSigner("first"), WithProposerServiceAccount())
		litter.Dump(res.EmulatorLog)
		assert.Contains(t, res.EmulatorLog[4], "0x01cf0e2f2f715450")
	})

	t.Run("Run simple tx with custom proposer", func(t *testing.T) {
		res := o.Tx("arguments", WithArg("test", "foo"), WithPayloadSigner("first"), WithProposer("account"))
		assert.Contains(t, res.EmulatorLog[4], "0x01cf0e2f2f715450")
	})

	t.Run("Fail when invalid proposer", func(t *testing.T) {
		res := o.Tx("arguments", WithArg("test", "foo"), WithPayloadSigner("first"), WithProposer("account2"))
		assert.ErrorContains(t, res.Err, "could not find account with name emulator-account2 in the configuration")
	})

	t.Run("Run linine tx", func(t *testing.T) {
		res := o.Tx(`
transaction(test:String) {
  prepare(acct: AuthAccount) {
    log(test)
 }
}
`, WithArg("test", "foo"), WithSignerServiceAccount())
		assert.NoError(t, res.Err)
	})

	t.Run("Run linine tx", func(t *testing.T) {
		res := o.Tx(`
transaction(test:UInt64) {
  prepare(acct: AuthAccount) {
    log(test)
 }
}
`, WithArg("test", uint64(1)), WithSignerServiceAccount())
		assert.NoError(t, res.Err)
	})

	t.Run("Run simple tx with custom signer", func(t *testing.T) {
		res := o.Tx("arguments", WithArg("test", "foo"), WithSigner("account"))
		assert.NoError(t, res.Err)
	})

	t.Run("Error on wrong signer name", func(t *testing.T) {
		res := o.Tx("arguments", WithArg("test", "foo"), WithSigner("account2"))
		assert.ErrorContains(t, res.Err, "could not find account with name emulator-account2")
	})

	t.Run("compose a function", func(t *testing.T) {
		serviceAccountTx := o.TxFN(WithSignerServiceAccount())
		res := serviceAccountTx("arguments", WithArg("test", "foo"))
		assert.NoError(t, res.Err)
	})

	t.Run("create function with name", func(t *testing.T) {
		argumentTx := o.TxFileNameFN("arguments", WithSignerServiceAccount())
		res := argumentTx(WithArg("test", "foo"))
		assert.NoError(t, res.Err)
	})

	t.Run("Should not allow varags builder arg with single element", func(t *testing.T) {
		res := o.Tx("arguments", WithArgs("test"))
		assert.ErrorContains(t, res.Err, "Please send in an even number of string : interface{} pairs")
	})

	t.Run("Should not allow varag with non string keys", func(t *testing.T) {
		res := o.Tx("arguments", WithArgs(1, "test"))
		assert.ErrorContains(t, res.Err, "even parameters in Args needs to be string")
	})

	t.Run("Arg, with cadence raw value", func(t *testing.T) {
		res := o.Tx("arguments", WithSignerServiceAccount(), WithArg("test", cadenceString("test")))
		assert.NoError(t, res.Err)
	})

	t.Run("date time arg", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:UFix64) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithArgDateTime("test", "July 29, 2021 08:00:00 AM", "America/New_York"), WithSignerServiceAccount())
		assert.NoError(t, res.Error)
	})

	t.Run("date time arg error", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:UFix64) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithArgDateTime("test", "July 29021 08:00:00 AM", "America/New_York"), WithSignerServiceAccount())
		assert.ErrorContains(t, res.Error, "cannot parse")
	})

	t.Run("Map args", func(t *testing.T) {
		res := o.Tx("arguments", WithSignerServiceAccount(), WithArgsMap(map[string]interface{}{"test": "test"}))
		assert.NoError(t, res.Err)
	})

	t.Run("Parse addresses should fail if not valid account name and hex", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:[Address]) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithAddresses("test", "bjartek"), WithSignerServiceAccount())
		assert.ErrorContains(t, res.Error, "bjartek is not an valid account name or an address")
	})

	t.Run("Parse array of addresses", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:[Address]) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithAddresses("test", "account", "45a1763c93006ca"), WithSignerServiceAccount())
		assert.Equal(t, "[0xf8d6e0586b0a20c7, 0x045a1763c93006ca]", fmt.Sprintf("%v", res.NamedArgs["test"]))
	})

	t.Run("Parse String to String map", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:{String:String}) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithArg("test", `{ "foo" : "bar"}`), WithSignerServiceAccount())
		assert.Equal(t, `{ "foo" : "bar"}`, fmt.Sprintf("%v", res.NamedArgs["test"]))
	})

	t.Run("Parse String to UFix64 map", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:{String:UFix64}) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithArg("test", `{ "foo" : 1.0}`), WithSignerServiceAccount())
		assert.Equal(t, `{ "foo" : 1.0}`, fmt.Sprintf("%v", res.NamedArgs["test"]))
	})

	t.Run("Error when parsing invalid address", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:Address) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithArg("test", "bjartek"), WithSignerServiceAccount())
		assert.ErrorContains(t, res.Error, "argument `test` with value `0xbjartek` is not expected type `Address`")

	})

	t.Run("Should set gas", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:Address) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithArg("test", "bjartek"), WithSignerServiceAccount(), WithMaxGas(100))

		assert.Equal(t, uint64(100), res.GasLimit)

	})

	t.Run("Should report error if invalid payload signer", func(t *testing.T) {
		res := o.Tx(`
transaction{
	prepare(acct: AuthAccount, user:AuthAccount) {

 }
}
`, WithSignerServiceAccount(), WithPayloadSigner("bjartek"))

		assert.Error(t, res.Err, "asd")

	})

	t.Run("ufix64", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:UFix64) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithArg("test", 1.0), WithSignerServiceAccount())
		assert.NoError(t, res.Error)
	})

	t.Run("add printer args", func(t *testing.T) {
		res := o.BuildInteraction(`
transaction(test:UFix64) {
  prepare(acct: AuthAccount) {

 }
}
`, "transaction", WithArg("test", 1.0), WithSignerServiceAccount(), WithPrintOptions(WithEmulatorLog()), WithPrintOptions(WithFullMeter()))
		assert.NoError(t, res.Error)
		assert.Equal(t, 2, len(*res.PrintOptions))
	})

	t.Run("add event filter", func(t *testing.T) {
		filter := OverflowEventFilter{
			"Deposit": []string{"id"},
		}

		res := o.BuildInteraction(`
transaction(test:UFix64) {
  prepare(acct: AuthAccount) {
 }
}
`, "transaction", WithArg("test", 1.0), WithSignerServiceAccount(), WithoutGlobalEventFilter(), WithEventsFilter(filter))
		assert.NoError(t, res.Error)
		assert.True(t, res.IgnoreGlobalEventFilters)
		assert.Equal(t, 1, len(res.EventFilter))
	})
}

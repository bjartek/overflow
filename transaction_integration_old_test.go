package overflow

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
 Tests must be in the same folder as flow.json with contracts and transactions/scripts in subdirectories in order for the path resolver to work correctly
*/
func TestTransactionIntegrationLegacy(t *testing.T) {
	logNumName := "A.f8d6e0586b0a20c7.Debug.LogNum"
	g := NewTestingEmulator().Start()
	g.Tx("mint_tokens", SignProposeAndPayAsServiceAccount(), Arg("recipient", "first"), Arg("amount", 1.0))

	t.Parallel()

	t.Run("fail on missing signer with run method", func(t *testing.T) {
		assert.PanicsWithError(t, "💩 You need to set the proposer signer", func() {
			g.TransactionFromFile("create_nft_collection").Run()
		})
	})

	t.Run("fail on missing signer", func(t *testing.T) {
		g.TransactionFromFile("create_nft_collection").
			Test(t).                                             //This method will return a TransactionResult that we can assert upon
			AssertFailure("You need to set the proposer signer") //we assert that there is a failure
	})

	t.Run("fail on wrong transaction name", func(t *testing.T) {
		g.TransactionFromFile("create_nf_collection").
			SignProposeAndPayAs("first").
			Test(t).                                                                                           //This method will return a TransactionResult that we can assert upon
			AssertFailure("Could not read interaction file from path=./transactions/create_nf_collection.cdc") //we assert that there is a failure
	})

	t.Run("Create NFT collection with different base path", func(t *testing.T) {
		g.TransactionFromFile("create_nft_collection").
			SignProposeAndPayAs("first").
			TransactionPath("./tx").
			Test(t).        //This method will return a TransactionResult that we can assert upon
			AssertSuccess() //Assert that there are no errors and that the transactions succeeds
	})

	t.Run("Mint tokens assert events", func(t *testing.T) {
		result := g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			ArgsFn(func(args *FlowArgumentsBuilder) {
				args.Account("first")
				args.UFix64(100.1)
			}).
			Test(t).
			AssertSuccess().
			AssertEventCount(6).                                                                                                                                            //assert the number of events returned
			AssertPartialEvent(NewTestEvent("A.0ae53cb6e3f42a79.FlowToken.TokensDeposited", map[string]interface{}{"amount": float64(100.1)})).                             //assert a given event, can also take multiple events if you like
			AssertEmitEventNameShortForm("FlowToken.TokensMinted").                                                                                                         //assert the name of a single event
			AssertEmitEventName("A.0ae53cb6e3f42a79.FlowToken.TokensMinted", "A.0ae53cb6e3f42a79.FlowToken.TokensDeposited", "A.0ae53cb6e3f42a79.FlowToken.MinterCreated"). //or assert more then one eventname in a go
			AssertEmitEvent(NewTestEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted", map[string]interface{}{"amount": float64(100.1)})).
			//assert a given event, can also take multiple events if you like
			AssertEmitEventJson("{\n  \"name\": \"A.0ae53cb6e3f42a79.FlowToken.MinterCreated\",\n  \"time\": \"1970-01-01T00:00:00Z\",\n  \"fields\": {\n    \"allowedAmount\": 100.1\n  }\n}") //assert a given event using json, can also take multiple events if you like

		assert.Equal(t, 1, len(result.Result.GetEventsWithName("A.0ae53cb6e3f42a79.FlowToken.TokensDeposited")))
		assert.Equal(t, 1, len(result.Result.GetEventsWithName("A.0ae53cb6e3f42a79.FlowToken.TokensDeposited")))

	})

	t.Run("Assert get id", func(t *testing.T) {
		result := g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`).
			SignProposeAndPayAs("first").
			Args(g.Arguments().UInt64(1)).
			Test(t).
			AssertSuccess()

		res, err := result.Result.GetIdFromEvent(logNumName, "id")
		assert.NoError(t, err)

		assert.Equal(t, uint64(1), result.GetIdFromEvent(logNumName, "id"))
		assert.Equal(t, uint64(1), res)
		assert.Equal(t, []uint64{1}, result.Result.GetIdsFromEvent(logNumName, "id"))

	})

	t.Run("run get id print all", func(t *testing.T) {
		result := g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`).
			SignProposeAndPayAs("first").
			Args(g.Arguments().UInt64(1)).
			RunGetIdFromEventPrintAll(logNumName, "id")

		assert.Equal(t, uint64(1), result)
	})

	t.Run("run get ids", func(t *testing.T) {
		result, err := g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`).
			SignProposeAndPayAs("first").
			Args(g.Arguments().UInt64(1)).
			RunGetIds(logNumName, "id")

		assert.NoError(t, err)
		assert.Equal(t, []uint64{1}, result)
	})

	t.Run("run get ids fail", func(t *testing.T) {
		_, err := g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`).
			SignProposeAndPayAs("first").
			RunGetIds(logNumName, "id")

		assert.ErrorContains(t, err, "entry point parameter count mismatch: expected 1, got 0")
	})

	t.Run("run get id print all panic on failed", func(t *testing.T) {

		assert.Panics(t, func() {
			g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`).SignProposeAndPayAs("first").RunGetIdFromEvent(logNumName, "id")
		})

	})

	t.Run("run get id print all panic on wrong field name", func(t *testing.T) {

		assert.Panics(t, func() {
			g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`).SignProposeAndPayAs("first").
				Args(g.Arguments().UInt64(1)).
				RunGetIdFromEvent(logNumName, "id2")
		})

	})

	t.Run("run get events with name", func(t *testing.T) {
		result := g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`).
			SignProposeAndPayAs("first").
			Args(g.Arguments().UInt64(1)).
			RunGetEventsWithName(logNumName)

		assert.Equal(t, 1, len(result))
	})

	t.Run("run get events with name or error", func(t *testing.T) {
		result, err := g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(id:UInt64) {
		  prepare(acct: AuthAccount) {
			  Debug.id(id) 
			} 
		}`).
			SignProposeAndPayAs("first").
			Args(g.Arguments().UInt64(1)).
			RunGetEventsWithNameOrError(logNumName)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
	})

	t.Run("Inline transaction with debug log", func(t *testing.T) {
		g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(message:String) {
		  prepare(acct: AuthAccount, account2: AuthAccount) {
			Debug.log(message) } }`).
			SignProposeAndPayAs("first").
			PayloadSigner("second").
			Args(g.Arguments().String("foobar")).
			Test(t).
			AssertSuccess().
			AssertDebugLog("foobar"). //assert that we have debug logged something. The assertion is contains so you do not need to write the entire debug log output if you do not like
			AssertComputationUsed(37).
			AssertEmulatorLog("Transaction submitted")

	})

	t.Run("Raw account argument", func(t *testing.T) {
		g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(user:Address) {
		  prepare(acct: AuthAccount) {
			Debug.log(user.toString())
		 }
		}`).
			SignProposeAndPayAsService().
			Args(g.Arguments().RawAccount("0x01cf0e2f2f715450")).
			Test(t).
			AssertSuccess().
			AssertDebugLog("0x01cf0e2f2f715450").
			AssertComputationLessThenOrEqual(40)
	})

	t.Run("transaction that should fail", func(t *testing.T) {
		g.Transaction(`
		import Debug from "../contracts/Debug.cdc"
		transaction(user:Address) {
		  prepare(acct: AuthAccount) {
			Debug.log(user.toStrig())
		 }
		}`).
			SignProposeAndPayAsService().
			Args(g.Arguments().
				RawAccount("0x1cf0e2f2f715450")).
			Test(t).
			AssertFailure("has no member `toStrig`") //assert failure with an error message. uses contains so you do not need to write entire message
	})

	t.Run("Assert print events", func(t *testing.T) {
		var str bytes.Buffer
		log.SetOutput(&str)
		defer log.SetOutput(os.Stdout)

		g.SimpleTxArgs("mint_tokens", "account", g.Arguments().Account("first").UFix64(100.0))
		assert.Contains(t, str.String(), "A.0ae53cb6e3f42a79.FlowToken.MinterCreated")
	})

	t.Run("Assert print events", func(t *testing.T) {
		var str bytes.Buffer
		log.SetOutput(&str)
		defer log.SetOutput(os.Stdout)

		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			Args(g.Arguments().
				Account("first").
				UFix64(100.0)).
			RunPrintEvents(map[string][]string{"A.0ae53cb6e3f42a79.FlowToken.TokensDeposited": {"to"}})

		assert.NotContains(t, str.String(), "0x1cf0e2f2f715450")
	})

	/*
		https://github.com/bjartek/overflow/issues/45
		t.Run("Meters test", func(t *testing.T) {
			res := g.TransactionFromFile("mint_tokens").
				SignProposeAndPayAsService().
				NamedArguments(map[string]string{
					"recipient": "first",
					"amount":    "100.0",
				}).
				Test(t).AssertSuccess()

			assert.Equal(t, 0, res.Result.Meter.Loops())
			assert.Equal(t, 15, res.Result.Meter.FunctionInvocations())
			assert.Equal(t, 42, res.Result.ComputationUsed)

		})
	*/

	t.Run("Named arguments wrong type", func(t *testing.T) {
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			NamedArguments(map[string]string{
				"recipient": "first",
				"amount":    "asd",
			}).
			Test(t).AssertFailure("argument `amount` is not expected type `UFix64`")
	})

	t.Run("Named arguments with string", func(t *testing.T) {
		g.TransactionFromFile("arguments").
			SignProposeAndPayAsService().
			NamedArguments(map[string]string{
				"test": "first",
			}).
			Test(t).AssertSuccess()
	})

	t.Run("Named arguments error if not all arguments", func(t *testing.T) {
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			NamedArguments(map[string]string{
				"recipient": "first",
			}).
			Test(t).AssertFailure("the following arguments where not present [amount]")
	})

	t.Run("Named arguments error if file not correct", func(t *testing.T) {
		g.TransactionFromFile("mint_tokens2").
			SignProposeAndPayAsService().
			NamedArguments(map[string]string{
				"recipient": "first",
			}).
			Test(t).AssertFailure("Could not read interaction file from path=./transactions/mint_tokens2.cdc")
	})
}

func TestFillUpSpace(t *testing.T) {
	o, err := OverflowTesting(WithFlowForNewUsers(0.0003))
	assert.NoError(t, err)

	result := o.GetFreeCapacity("first")
	assert.Equal(t, 129123, result)
	o.FillUpStorage("first")
	assert.NoError(t, o.Error)

	result2 := o.GetFreeCapacity("first")
	assert.LessOrEqual(t, result2, 100)

}

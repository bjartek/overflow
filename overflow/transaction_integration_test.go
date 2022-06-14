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
func TestTransactionIntegration(t *testing.T) {
	logNumName := "A.f8d6e0586b0a20c7.Debug.LogNum"
	g := NewTestingEmulator().Start()
	t.Parallel()

	t.Run("fail on missing signer with run method", func(t *testing.T) {
		assert.PanicsWithError(t, "💩 You need to set the main signer", func() {
			g.TransactionFromFile("create_nft_collection").Run()
		})
	})

	t.Run("fail on missing signer", func(t *testing.T) {
		g.TransactionFromFile("create_nft_collection").
			Test(t).                                         //This method will return a TransactionResult that we can assert upon
			AssertFailure("You need to set the main signer") //we assert that there is a failure
	})

	t.Run("fail on wrong transaction name", func(t *testing.T) {
		g.TransactionFromFile("create_nf_collection").
			SignProposeAndPayAs("first").
			Test(t).                                                                                           //This method will return a TransactionResult that we can assert upon
			AssertFailure("Could not read transaction file from path=./transactions/create_nf_collection.cdc") //we assert that there is a failure
	})

	t.Run("Create NFT collection with differnt base path", func(t *testing.T) {
		g.TransactionFromFile("create_nft_collection").
			SignProposeAndPayAs("first").
			TransactionPath("./tx").
			Test(t).         //This method will return a TransactionResult that we can assert upon
			AssertSuccess(). //Assert that there are no errors and that the transactions succeeds
			AssertNoEvents() //Assert that we did not emit any events.
	})

	t.Run("Mint tokens assert events", func(t *testing.T) {
		result := g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			ArgsFn(func(args *FlowArgumentsBuilder) {
				args.Account("first")
				args.UFix64(100.0)
			}).
			Test(t).
			AssertSuccess().
			AssertEventCount(3).                                                                                                                                                                           //assert the number of events returned
			AssertPartialEvent(NewTestEvent("A.0ae53cb6e3f42a79.FlowToken.TokensDeposited", map[string]interface{}{"amount": "100.00000000"})).                                                            //assert a given event, can also take multiple events if you like
			AssertEmitEventNameShortForm("FlowToken.TokensMinted").                                                                                                                                        //assert the name of a single event
			AssertEmitEventName("A.0ae53cb6e3f42a79.FlowToken.TokensMinted", "A.0ae53cb6e3f42a79.FlowToken.TokensDeposited", "A.0ae53cb6e3f42a79.FlowToken.MinterCreated").                                //or assert more then one eventname in a go
			AssertEmitEvent(NewTestEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted", map[string]interface{}{"amount": "100.00000000"})).                                                                  //assert a given event, can also take multiple events if you like
			AssertEmitEventJson("{\n  \"name\": \"A.0ae53cb6e3f42a79.FlowToken.MinterCreated\",\n  \"time\": \"1970-01-01T00:00:00Z\",\n  \"fields\": {\n    \"allowedAmount\": \"100.00000000\"\n  }\n}") //assert a given event using json, can also take multiple events if you like

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

		assert.Equal(t, uint64(1), result.Result.GetIdFromEvent(logNumName, "id"))

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
			AssertComputationUsed(5).
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
			AssertComputationLessThenOrEqual(10)
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

	t.Run("Named arguments", func(t *testing.T) {
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			NamedArguments(map[string]string{
				"recipient": "first",
				"amount":    "100.0",
			}).
			Test(t).AssertSuccess()

	})

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
			Test(t).AssertFailure("Could not read transaction file from path=./transactions/mint_tokens2.cdc")
	})

	t.Run("Get free capacity", func(t *testing.T) {
		result := g.GetFreeCapacity("second")
		assert.Equal(t, 99104, result)
		g.FillUpStorage("second")

		result2 := g.GetFreeCapacity("second")
		assert.Equal(t, 0, result2)
	})

}

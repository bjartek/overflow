package main

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bjartek/overflow/overflow"
)

/*
 Tests must be in the same folder as flow.json with contracts and transactions/scripts in subdirectories in order for the path resolver to work correctly
*/
func TestTransaction(t *testing.T) {
	g := overflow.NewTestingEmulator().Start()
	t.Parallel()

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
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			ArgsFn(func(args *overflow.FlowArgumentsBuilder) {
				args.Account("first")
				args.UFix64(100.0)
			}).
			Test(t).
			AssertSuccess().
			AssertEventCount(3).                                                                                                                                                                           //assert the number of events returned
			AssertPartialEvent(overflow.NewTestEvent("A.0ae53cb6e3f42a79.FlowToken.TokensDeposited", map[string]interface{}{"amount": "100.00000000"})).                                                   //assert a given event, can also take multiple events if you like
			AssertEmitEventNameShortForm("FlowToken.TokensMinted").                                                                                                                                        //assert the name of a single event
			AssertEmitEventName("A.0ae53cb6e3f42a79.FlowToken.TokensMinted", "A.0ae53cb6e3f42a79.FlowToken.TokensDeposited", "A.0ae53cb6e3f42a79.FlowToken.MinterCreated").                                //or assert more then one eventname in a go
			AssertEmitEvent(overflow.NewTestEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted", map[string]interface{}{"amount": "100.00000000"})).                                                         //assert a given event, can also take multiple events if you like
			AssertEmitEventJson("{\n  \"name\": \"A.0ae53cb6e3f42a79.FlowToken.MinterCreated\",\n  \"time\": \"1970-01-01T00:00:00Z\",\n  \"fields\": {\n    \"allowedAmount\": \"100.00000000\"\n  }\n}") //assert a given event using json, can also take multiple events if you like

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
			Args(g.Arguments().RawAccount("0x1cf0e2f2f715450")).
			Test(t).
			AssertSuccess().
			AssertDebugLog("0x1cf0e2f2f715450").
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

		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			ArgsV(g.Arguments().Account("first").UFix64(100.0).Build()).
			RunPrintEventsFull()

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
}

package main

import (
	"log"

	"github.com/bjartek/overflow/overflow"
	"github.com/onflow/cadence"
)

func main() {

	//This method starts an in memory flow emulator
	// - it then looks at all the contracts in the deployment block for emulator and deploys them
	// - then it looks at all the accounts that does not have contracts in them and create those accounts. These can be used as stakeholders in your "storyline" below.
	// - when referencing accounts in the "storyline" below note that the default option is to prepened the network to the account name, This is done so that it is easy to run a storyline against emulator, tesnet and mainnet. This can be disabled with the `DoNotPrependNetworkToAccountNames` method on the overflow builder

	// Note that if you want this to run against an already running emulator and not an embedded one run the script with OVERFLOW_ENV=emulator

	g := overflow.NewOverflow().Start()

	structValue := cadence.Struct{
		Fields: []cadence.Value{cadence.String("baz")},
		StructType: &cadence.StructType{
			QualifiedIdentifier: "A.f8d6e0586b0a20c7.Debug.Foo",
			Fields: []cadence.Field{{
				Identifier: "bar",
				Type:       cadence.StringType{},
			}},
		},
	}

	g.Transaction(`
import Debug from "../contracts/Debug.cdc"

transaction(value:Debug.Foo) {
  prepare(acct: AuthAccount) {
	Debug.log(value.bar)
 }
}`).SignProposeAndPayAs("first").Args(g.Arguments().Argument(structValue)).RunPrintEventsFull()

	//this first transaction will setup a NFTCollection for the user "emulator-first".
	// transactions are looked up in the `transactions` folder.
	//if we change the initialization of overflow to testnet above the account used here would be "testnet-first".
	// finally we run the transaction and print all the events, there are several convenience methods to filter out fields from events of not print them at all if you like.
	g.TransactionFromFile("create_nft_collection").SignProposeAndPayAs("first").RunPrintEventsFull()

	//the second transaction show how you can call a transaction with an argument. In this case we send a string to the transactions
	g.TransactionFromFile("arguments").SignProposeAndPayAs("first").Args(g.Arguments().String("argument1")).RunPrintEventsFull()

	//it is possible to send an accounts address as argument to a script using a convenience function `Account`. Network is prefixed here as well
	g.TransactionFromFile("argumentsWithAccount").SignProposeAndPayAs("first").Args(g.Arguments().Account("second")).RunPrintEventsFull()

	//This transactions shows an example of signing the main envelope with the "first" user and the paylod with the "second" user.
	g.TransactionFromFile("signWithMultipleAccounts").SignProposeAndPayAs("first").PayloadSigner("second").Args(g.Arguments().String("asserts.go")).RunPrintEventsFull()

	//Running a script from a file is almost like running a transaction.
	g.ScriptFromFile("test").Args(g.Arguments().Account("second")).Run()

	//In this transaction we actually do some meaningful work. We mint 10 flowTokens into the account of user first. Note that this method will not work on mainnet or testnet. If you want tokens on testnet use the faucet or transfer from one account to another
	g.TransactionFromFile("mint_tokens").SignProposeAndPayAsService().Args(g.Arguments().Account("first").UFix64(10.0)).RunPrintEventsFull()

	//If you do not want to store a script in a file you can use a inline representation with go multiline strings
	g.Script(`
pub fun main(account: Address): String {
    return getAccount(account).address.toString()
}`).Args(g.Arguments().Account("second")).Run()

	//The same is also possible for a transaction. Also note the handy Debug contracts log method that allow you to assert some output from a transaction other then an event.
	g.Transaction(`
import Debug from "../contracts/Debug.cdc"
transaction(value:String) {
  prepare(acct: AuthAccount) {
	Debug.log(value)
 }
}`).SignProposeAndPayAs("first").Args(g.Arguments().String("foobar")).RunPrintEventsFull()

	//Run script that returns
	result := g.ScriptFromFile("test").Args(g.Arguments().Account("second")).RunFailOnError()
	log.Printf("Script returned %s", result)

}

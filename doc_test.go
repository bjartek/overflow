package overflow_test

// importing overflow using "." will yield a cleaner DSL
import (
	"fmt"
	"time"

	. "github.com/bjartek/overflow"
)

func Example() {

	//in order to start overflow use the Overflow function
	//it can be customized with lots of OverflowOption
	Overflow()
	//Output:
	//📜 deploy contracts NonFungibleToken, Debug
	//🧑 Created account: emulator-first with address: 01cf0e2f2f715450 with flow: 10.00
	//🧑 Created account: emulator-second with address: 179b6b1cb6755e31 with flow: 10.00
}

func ExampleOverflowState_Tx() {
	o := Overflow()

	// start the Tx DSL with the name of the transactions file, by default this
	// is in the `transactions` folder in your root dit
	o.Tx("arguments",
		//Customize the Transaction by sending in more InteractionOptions,
		//at minimum you need to set Signer and Args if any
		SignProposeAndPayAs("first"),
		//Arguments are always passed by name in the DSL builder, order does not matter
		Arg("test", "overflow ftw!"),
	)
}

func ExampleOverflowState_Tx_inline() {
	o := Overflow()

	//The Tx dsl can also contain an inline transaction
	o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(message:String) {
		  prepare(acct: AuthAccount) {
				Debug.log(message) 
			} 
		}`,
		SignProposeAndPayAs("first"),
		Arg("message", "overflow ftw!"),
	)
}

func ExampleOverflowState_Tx_multisign() {
	o := Overflow()

	//The Tx dsl can also contain an inline transaction
	o.Tx(`
		transaction {
			prepare(acct: AuthAccount, acct2: AuthAccount) {
			  //aact here is first
				//acct2 here is second
			} 
		}`,
		SignProposeAndPayAs("first"),
		PayloadSigner("second"),
	)

}

func ExampleOverflowState_Script() {
	o := Overflow()

	// the other major interaction you can run on Flow is a script, it uses the script DSL.
	// Start it by specifying the script name from `scripts` folder
	o.Script("test",
		// the test script requires an address as arguments, Overflow is smart enough that it
		// sees this and knows that there is an account for the emulator network called
		// `emulator-first` so it will insert that address as the argument.
		// If you change the network to testnet/mainnet later and name your stakholders
		// accordingly it will just work
		Arg("account", "first"),
	)
}

func ExampleOverflowState_Script_inline() {
	o := Overflow()

	//Script can be run inline
	o.Script(`
pub fun main(account: Address): String {
    return getAccount(account).address.toString()
}`,
		Arg("account", "first"),
	)
}

func ExampleOverflowState_FetchEvents() {
	o := Overflow()

	for {
		events, err := o.FetchEvents(
			TrackProgressIn("minted_tokens"),
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
		)
		if err != nil {
			panic(err)
		}

		if len(events) == 0 {
			//here you can specify how long you will wait between polls
			time.Sleep(10 * time.Second)
		}

		// do something with events, like sending them to discord/twitter or index in a database
		fmt.Println(events)
	}
}

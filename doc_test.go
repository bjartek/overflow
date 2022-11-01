package overflow_test

// importing overflow using "." will yield a cleaner DSL
import (
	. "github.com/bjartek/overflow"
)

var docOptions = WithGlobalPrintOptions(WithoutId())

func Example() {

	//in order to start overflow use the Overflow function
	//it can be customized with lots of OverflowOption
	Overflow()
	//Output:
	//🧑 Created account: emulator-first with address: 01cf0e2f2f715450 with flow: 10.00
	//🧑 Created account: emulator-second with address: 179b6b1cb6755e31 with flow: 10.00
	//📜 deploy contracts NonFungibleToken, Debug, MetadataViews
}

func ExampleOverflowState_Tx() {
	o := Overflow(docOptions)

	// start the Tx DSL with the name of the transactions file, by default this
	// is in the `transactions` folder in your root dit
	o.Tx("arguments",
		//Customize the Transaction by sending in more InteractionOptions,
		//at minimum you need to set Signer and Args if any
		WithSigner("first"),
		//Arguments are always passed by name in the DSL builder, order does not matter
		WithArg("test", "overflow ftw!"),
	)
	//Output:
	//🧑 Created account: emulator-first with address: 01cf0e2f2f715450 with flow: 10.00
	//🧑 Created account: emulator-second with address: 179b6b1cb6755e31 with flow: 10.00
	//📜 deploy contracts NonFungibleToken, Debug, MetadataViews
	//👌 Tx:arguments fee:0.00000100 gas:0
	//
}

func ExampleOverflowState_Tx_inline() {
	o := Overflow(docOptions)

	//The Tx dsl can also contain an inline transaction
	o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction(message:String) {
		  prepare(acct: AuthAccount) {
				Debug.log(message) 
			} 
		}`,
		WithSigner("first"),
		WithArg("message", "overflow ftw!"),
	)
	//Output:
	//🧑 Created account: emulator-first with address: 01cf0e2f2f715450 with flow: 10.00
	//🧑 Created account: emulator-second with address: 179b6b1cb6755e31 with flow: 10.00
	//📜 deploy contracts NonFungibleToken, Debug, MetadataViews
	//👌 Tx: fee:0.00000134 gas:7
	//=== Events ===
	//A.f8d6e0586b0a20c7.Debug.Log
	//   msg -> overflow ftw!
}

func ExampleOverflowState_Tx_multisign() {
	o := Overflow(docOptions)

	//The Tx dsl supports multiple signers, note that the mainSigner is the last account
	o.Tx(`
		import Debug from "../contracts/Debug.cdc"
		transaction {
			prepare(acct: AuthAccount, acct2: AuthAccount) {
				Debug.log("acct:".concat(acct.address.toString()))
				Debug.log("acct2:".concat(acct2.address.toString()))
			} 
		}`,
		WithSigner("first"),
		WithPayloadSigner("second"),
	)

	//Output:
	//🧑 Created account: emulator-first with address: 01cf0e2f2f715450 with flow: 10.00
	//🧑 Created account: emulator-second with address: 179b6b1cb6755e31 with flow: 10.00
	//📜 deploy contracts NonFungibleToken, Debug, MetadataViews
	//👌 Tx: fee:0.00000134 gas:7
	//=== Events ===
	//A.f8d6e0586b0a20c7.Debug.Log
	//   msg -> acct:0x179b6b1cb6755e31
	//A.f8d6e0586b0a20c7.Debug.Log
	//   msg -> acct2:0x01cf0e2f2f715450
	//
}

func ExampleOverflowState_Script() {
	o := Overflow(docOptions)

	// the other major interaction you can run on Flow is a script, it uses the script DSL.
	// Start it by specifying the script name from `scripts` folder
	o.Script("test",
		// the test script requires an address as arguments, Overflow is smart enough that it
		// sees this and knows that there is an account for the emulator network called
		// `emulator-first` so it will insert that address as the argument.
		// If you change the network to testnet/mainnet later and name your stakholders
		// accordingly it will just work
		WithArg("account", "first"),
	)
	//Output:
	//🧑 Created account: emulator-first with address: 01cf0e2f2f715450 with flow: 10.00
	//🧑 Created account: emulator-second with address: 179b6b1cb6755e31 with flow: 10.00
	//📜 deploy contracts NonFungibleToken, Debug, MetadataViews
	//⭐ Script test run result:"0x01cf0e2f2f715450"
}

func ExampleOverflowState_Script_inline() {

	o := Overflow(docOptions)

	//Script can be run inline
	o.Script(`
pub fun main(account: Address): String {
    return getAccount(account).address.toString()
}`,
		WithArg("account", "first"),
		WithName("get_address"),
	)
	//Output:
	//🧑 Created account: emulator-first with address: 01cf0e2f2f715450 with flow: 10.00
	//🧑 Created account: emulator-second with address: 179b6b1cb6755e31 with flow: 10.00
	//📜 deploy contracts NonFungibleToken, Debug, MetadataViews
	//⭐ Script get_address run result:"0x01cf0e2f2f715450"
}

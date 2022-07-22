package overflow_test

// importing overflow using "." will yield a cleaner DSL
import . "github.com/bjartek/overflow"

func Example() {

	// create a new overflow emulator that will panic if scripts/transactions fail and print output
	o := Overflow(
		StopOnError(),
		PrintInteractionResults(),
		// here you can send in more options to customize the way Overflow is started
	)

	// start the Tx DSL with the name of the transactions file, by default this is in the `transactions` folder in your root dit
	o.Tx("arguments",
		//Customize the Transaction by sending in more InteractionOptions, at minimum you need to set Signer and Args if any
		SignProposeAndPayAs("first"),
		//Arguments are always passed by name in the DSL builder, order does not matter
		Arg("test", "overflow ftw!"),
	)
	// Output: 👌 Tx:arguments computation:?? loops:? statements:? invocations:? id:?
	// the standard output if you print results continas the name of the transaction, id and computation information if run on emulator

	// the other major interaction you can run on Flow is a script, it uses the script DSL. Start it by specifying the script name from `scripts` folder
	o.Script("test",
		//the test script requires an address as arguments, Overflow is smart enough that it sees this and knows that there is an account for the emulator network called `emulator-first` so it will insert that address as the argument. If you change the network to testnet/mainnet later and name your stakholders accordingly it will just work
		Arg("account", "first"),
	)
	// Output: ⭐ Script test result:??
	// will print out the name of the script and the result as json

	// This is just the simples example of an interaction using overflow but it has many hidden gems!
}

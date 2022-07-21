[![Coverage Status](https://coveralls.io/repos/github/bjartek/overflow/badge.svg?branch=main)](https://coveralls.io/github/bjartek/overflow?branch=main) [![ci](https://github.com/bjartek/overflow/actions/workflows/ci.yaml/badge.svg)](https://github.com/bjartek/overflow/actions/workflows/ci.yaml)

# Overflow

A DSL written in golang to be used in tests or to run a `story` of interactions against either an local emulator, testnet, mainnet or an in memory instance of the flow-emulator.

Use case scenarios include:
 - demo purposes
 - integration testing of combinations of scripts/transactions
 - batch jobs.

For a standalone example on how overflow can be used look at https://github.com/bjartek/flow-nft-overflow it has both an interactive demo and unit tests for an example NFT contract. 

## Main features

- Uses a shared golang builder pattern for almost all interactions. 
- Well documented source code
- supports all variants of multi-sign, Two authorizers, proposer and payer can be different. 
- when refering to an account/address you can use the logical name for that stakeholder defined in flow json. the same stakeholder IE admin can have different addresses on each network
- can be run in embedded in memory mode that will start emulator, deploy contracts, create stakeholders and run interactions (scripts/transactions) against this embedded system and then stop it when it ends. 
- has a DSL to fetch Events and optionally store progress in a file. This can be chained into indexers/crawlers/notification services. 
- all interactions can be specified inline as well as from files
- transform all interactions into a NPM module that can be published for the frontend to use. this json file that is generate has the option to filter out certain interactions and to strip away network suffixes if you have multiple local interactions that should map to the same logical name in the client for each network
- the interaction (script/tx) dsl has a rich set of assertions 
- arguments to interactions are all _named_ that is the same name in that is in the argument must be used with the `Arg("name", "value")` builder. The `value` in this example can be either a primitive go value or a `cadence.Value`. 

## Gotchas

- When specifying extra accounts that are created on emulator they are created in alphabetical order, the addresses the emulator assign is always fixed.
- tldr; Name your stakeholder accounts in alphabetical order, we suggest admin, bob, charlie, demi, eddie
- When writing integration tests, tests must be in the same folder as flow.json
with contracts and transactions/scripts in subdirectories in order for the path resolver
to work correctly

## Resources

- Check [other codebases](https://github.com/bjartek/overflow/network/dependents) that use this project
- Feel free to ask questions to @bjartek in the Overflow Discord. https://discord.gg/t6GEtHnWFh

## Usage

First create a project directory, initialize the go module and install `overflow`:

```sh
mkdir test-overflow && cd test-overflow
flow init
go mod init example.com/test-overflow
go get github.com/bjartek/overflow
```

Then create a task file:

```sh
touch tasks/main.go
```

In that task file, you can then import `overflow` and use it to your convenience, for example:

```go
package main

import (
    "fmt"

    //if you imports this with .  you do not have to repeat overflow everywhere 
    . "github.com/bjartek/overflow"
)

func main() {

	//start an in memory emulator by default
	o := Overflow()
	
	//the Tx DSL runs an transaction
	o.Tx("name_of_transaction", SignProposeAndPayAs("bob"), Arg("name", "bob")).Print()
	
	//Run a script/get interaction against the same in memory chain
	o.Script("name_of_script", Arg("name", "bob")).Print()
}
```

Then you can run

```sh
go run ./tasks/main.go
```

This is a minimal example that run a transction and a script, but from there you can branch out.

The following env vars are supported
 - OVERFLOW_ENV : set the environment to run against "emulator|embedded|testnet|mainnet|testing" (embedded is standard)
 - OVEFFLOW_CONTINUE: if you do not want overflow to deploy contracts and accounts on emulator you can set this to true
 - OVERFLOW_LOGGING: Set this to 0-4 to get increasing log

## Credits

This project is the successor of https://github.com/bjartek/go-with-the-flow
The v0 version of the code with a set of apis that is now deprecated is in https://github.com/bjartek/overflow/tree/v0

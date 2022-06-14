[![Coverage Status](https://coveralls.io/repos/github/bjartek/overflow/badge.svg?branch=main)](https://coveralls.io/github/bjartek/overflow?branch=main) [![ci](https://github.com/bjartek/overflow/actions/workflows/ci.yaml/badge.svg)](https://github.com/bjartek/overflow/actions/workflows/ci.yaml)



# Overflow

> Tooling to help develop application on the the Flow Blockchain

Set of go scripts to make it easer to run a story consisting of creating accounts,
deploying contracts, executing transactions and running scripts on the Flow Blockchain.
These go scripts also make writing integration tests of your smart contracts much easier.


## Information

### Main features

- Create a single go file that will start emulator, deploy contracts, create accounts and run scripts and transactions. see `examples/demo/main.go`
- Fetch events, store progress in a file and send results to Discord. see `examples/event/main.go`
- Support inline scripts if you do not want to store everything in a file when testing
- Supports writing tests against transactions and scripts with some limitations on how to implement them.
- Asserts to make it easier to use the library in writing tests see `examples/transaction_test.go` for examples

### Gotchas

- When specifying extra accounts that are created on emulator they are created in alphabetical order, the addresses the emulator assign is always fixed.
- tldr; Name your stakeholder accounts in alphabetical order
- When writing integration tests, tests must be in the same folder as flow.json
with contracts and transactions/scripts in subdirectories in order for the path resolver
to work correctly

## Resources

- Check [other codebases](https://github.com/bjartek/overflow/network/dependents) that use this project
- Feel free to ask questions to @bjartek in the Flow Discord.

## Usage

First create a project directory, initialize the go module and install `overflow`:

```sh
mkdir test-overflow && cd test-overflow
flow init
go mod init example.com/test-overflow
go get github.com/bjartek/overflow/overflow
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

    "github.com/bjartek/overflow/overflow"
)

func main() {
    o := overflow.NewOverflow().Start()
    fmt.Printf("%v", o.State.Accounts())
}
```

Then you can run

```sh
go run ./tasks/main.go
```

This is a minimal example that only prints accounts, but from there you can branch out.

The following env vars are supported
 - OVERFLOW_ENV : set the environment to run against "emulator|embedded|testnet|mainnet" (embedded is standard)
 - OVEFFLOW_CONTINUE: if you do not want overflow to deploy contracts and accounts on emulator you can set this to true
 - OVERFLOW_LOGGING: Set this to 0-4 to get increasing log

## Credits

This project is the successor of https://github.com/bjartek/go-with-the-flow

[![Coverage Status](https://coveralls.io/repos/github/bjartek/overflow/badge.svg?branch=main)](https://coveralls.io/github/bjartek/overflow?branch=main) [![ci](https://github.com/bjartek/overflow/actions/workflows/ci.yml/badge.svg)](https://github.com/bjartek/overflow/actions/workflows/ci.yml)

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

### Note on v2

v2 of GoWithTheFlow removed a lot of the code in favor of `flowkit` in the flow-cli. Some of the code from here was
contributed by me into flow-cli like the goroutine based event fetcher.

Breaking changes between v1 and v2:

- v1 had a config section for discord webhooks. That has been removed since the flow-cli will remove extra config things in flow.json. Store the webhook url in an env variable and use it as argument when creating the DiscordWebhook struct.

Special thanks to @sideninja for helping me get my changes into flow-cli. and for jayShen that helped with fixing some issues!

## Resources

- Run the demo example in this project with `cd example && make`. The emulator will be run in memory.
- Check [other codebases](https://github.com/bjartek/go-with-the-flow/network/dependents?package_id=UGFja2FnZS0yMjc1NjE0OTAz) that use this project
- Feel free to ask questions to @bjartek in the Flow Discord.

## Usage

First create a project directory, initialize the go module and install `go-with-the-flow`:

```sh
mkdir test-gwtf && cd test-gtwf
flow init
go mod init example.com/test-gwtf
go get github.com/bjartek/go-with-the-flow/v2/gwtf
```

Then create a task file:

```sh
touch tasks/main.go
```

In that task file, you can then import `go-with-the-flow` and use it to your convenience, for example:

```go
package main

import (
    "fmt"

    "github.com/bjartek/go-with-the-flow/v2/gwtf"
)

func main() {
    g := gwtf.NewGoWithTheFlowInMemoryEmulator()
    fmt.Printf("%v", g.State.Accounts())
}
```

Then you can run

```sh
go run ./tasks/main.go
```

This is a minimal example that only prints accounts, but from there you can branch out.

## Credits

This project is the successor of https://github.com/bjartek/go-with-the-flow

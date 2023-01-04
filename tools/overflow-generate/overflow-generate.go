package main

import (
	"fmt"
	"os"

	. "github.com/bjartek/overflow"
)

func main() {

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Call with <network> <pathToInteraction> or just <pathToInteraction>")
		os.Exit(1)

	}
	network := "mainnet"
	interaction := args[0]
	if len(args) > 1 {
		network = args[0]
		interaction = args[1]
	}

	o := Overflow(WithNetwork(network))

	stub, err := o.GenerateStub(network, interaction, true)
	if err != nil {
		fmt.Printf("Could not genearte stub for network %s interaction %s error:%v", network, interaction, err)
		os.Exit(1)
	}

	fmt.Println(stub)
}

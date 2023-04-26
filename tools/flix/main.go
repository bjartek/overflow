package main

import (
	"fmt"
	"os"

	"github.com/bjartek/overflow"
	"github.com/sanity-io/litter"
)

func main() {
	file := os.Args[1]
	argsWithoutProg := os.Args[2:]

	network := "testnet"
	flix, err := overflow.ReadFileIntoStruct[overflow.FlowInteractionTemplate](fmt.Sprintf("%s", file))
	if err != nil {
		panic(err)
	}

	code := flix.Data.ResolvedCadence(network)
	o := overflow.Overflow(overflow.WithNetwork(network), overflow.WithGlobalPrintOptions())

	args := []overflow.OverflowInteractionOption{}
	if flix.IsTransaction() {
		args = append(args, overflow.WithSigner(os.Args[2]))
		argsWithoutProg = os.Args[3:]
	}

	if len(argsWithoutProg) != len(flix.Data.Arguments) {
		litter.Dump(flix.Data.Arguments)
		os.Exit(1)
	}

	for _, arg := range flix.Data.Arguments {
		value := argsWithoutProg[arg.Index]
		args = append(args, overflow.WithArg(arg.Key, value))
	}
	if flix.Data.Type == "transaction" {
		o.Tx(code, args...)
	} else {
		o.Script(code, args...)
	}
}

package main

import (
	"fmt"

	"github.com/bjartek/overflow/overflow"
)

func main() {
	g := overflow.NewOverflow().ExistingEmulator().Start()

	res, err := g.ParseAllWithConfig(false, []string{}, []string{})
	if err != nil {
		panic(err)
	}

	contracts := res.Networks["mainnet"].Contracts

	for _, contract := range *contracts {
		fmt.Println(contract)
	}

}

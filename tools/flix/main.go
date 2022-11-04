package main

import (
	"github.com/bjartek/overflow"
	"github.com/sanity-io/litter"
)

func main() {

	flix, err := overflow.ReadFileIntoStruct[overflow.FlowInteractionTemplate]("flix.json")
	if err != nil {
		panic(err)
	}
	litter.Dump(flix)
}

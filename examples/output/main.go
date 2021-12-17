package main

import (
	"encoding/json"
	"fmt"

	"github.com/bjartek/overflow/overflow"
)

func main() {
	g := overflow.NewOverflow().ExistingEmulator().Start()

	res, err := g.ParseAll()
	if err != nil {
		panic(err)
	}

	file, _ := json.MarshalIndent(res, "", "   ")
	fmt.Println(string(file))
	//	_ = ioutil.WriteFile("test.json", file, 0644)
	//	fmt.Println("outputted to test.json")
}

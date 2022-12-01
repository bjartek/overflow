package main

import (
	"fmt"
	"io"
	"os"

	"github.com/bjartek/overflow"
	jsoncdc "github.com/onflow/cadence/encoding/json"
)

func main() {
	_, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	value, err := jsoncdc.Decode(nil, bytes)
	if err != nil {
		panic(err)
	}
	output, err := overflow.CadenceValueToJsonString(value)
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}

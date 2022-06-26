package overflow

import (
	"encoding/hex"
	"strings"

	"github.com/onflow/cadence"
)

func CadenceString(input string) cadence.String {
	value, err := cadence.NewString(input)
	if err != nil {
		panic(err)
	}
	return value
}

// HexToAddress converts a hex string to an Address.
func HexToAddress(h string) (*cadence.Address, error) {
	trimmed := strings.TrimPrefix(h, "0x")
	if len(trimmed)%2 == 1 {
		trimmed = "0" + trimmed
	}
	b, err := hex.DecodeString(trimmed)
	if err != nil {
		return nil, err

	}
	address := cadence.BytesToAddress(b)
	return &address, nil
}

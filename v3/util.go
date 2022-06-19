package v3

import (
	"encoding/hex"
	"strings"

	"github.com/onflow/cadence"
)

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

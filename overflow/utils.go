package overflow

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/araddon/dateparse"
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

func parseTime(timeString string, location string) (string, error) {
	loc, err := time.LoadLocation(location)
	if err != nil {
		return "", err
	}
	time.Local = loc
	t, err := dateparse.ParseLocal(timeString)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.0", t.Unix()), nil
}

func getAndUnquoteStringAsPointer(value cadence.Value) *string {
	result, err := strconv.Unquote(value.String())
	if err != nil {
		result = value.String()
	}

	if result == "" {
		return nil
	}
	return &result
}

package overflow

import (
	"github.com/onflow/cadence"
)

const (
	fixedPointrecisionMultiple = 100000000
)

func CadenceString(input string) cadence.String {
	value, err := cadence.NewString(input)
	if err != nil {
		panic(err)
	}
	return value
}

func CadenceValueToGoValue(input cadence.Value) (output interface{}) {
	val := input.ToGoValue()
	switch input.(type) {
	// TODO: can these be handled together?
	case cadence.UFix64:
		output = float64(val.(uint64)) / fixedPointrecisionMultiple
		break
	case cadence.Fix64:
		output = float64(val.(int64)) / fixedPointrecisionMultiple
		break
	default:
		output = val
	}
	return
}

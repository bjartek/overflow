package overflow

import "github.com/onflow/cadence"

//TODO: Remove this
func CadenceString(input string) cadence.String {
	value, err := cadence.NewString(input)
	if err != nil {
		panic(err)
	}
	return value
}

package overflow

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/onflow/cadence"
)

type Cadencetest struct {
	name  string
	value cadence.Value
}

func TestCadenceValueToInterfaceCompact(t *testing.T) {

	foo := CadenceString("foo")
	bar := CadenceString("bar")
	emptyString := CadenceString("")

	emptyStrct := cadence.Struct{
		Fields: []cadence.Value{emptyString},
		StructType: &cadence.StructType{
			Fields: []cadence.Field{{
				Identifier: "foo",
				Type:       cadence.StringType{},
			}},
		},
	}
	strct := cadence.Struct{
		Fields: []cadence.Value{bar},
		StructType: &cadence.StructType{
			Fields: []cadence.Field{{
				Identifier: "foo",
				Type:       cadence.StringType{},
			}},
		},
	}
	dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: foo, Value: bar}})
	testCases := []Cadencetest{
		{"nil", nil},
		{"None", cadence.NewOptional(nil)},
		{"Some(string)", cadence.NewOptional(foo)},
		{"Some(uint64)", cadence.NewOptional(cadence.NewUInt64(42))},
		{"uint64", cadence.NewUInt64(42)},
		{"string array", cadence.NewArray([]cadence.Value{foo, bar})},
		{"empty array", cadence.NewArray([]cadence.Value{emptyString})},
		{"string array ignore empty", cadence.NewArray([]cadence.Value{foo, emptyString, bar})},
		{"dictionary", dict},
		{"dictionary_ignore_empty_value", cadence.NewDictionary([]cadence.KeyValuePair{{Key: foo, Value: emptyString}})},
		{"dictionary_with_subdict", cadence.NewDictionary([]cadence.KeyValuePair{{Key: bar, Value: dict}})},
		{"struct", strct},
		{"empty struct", emptyStrct},
	}

	for _, tc := range testCases {
		//		t.Parallel()

		t.Run(tc.name, func(t *testing.T) {
			value := CadenceValueToInterfaceCompact(tc.value)
			autogold.Equal(t, value)
		})
	}
}

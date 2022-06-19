package overflow

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
)

type Cadencetest struct {
	name   string
	value  cadence.Value
	result interface{}
}

func TestCadenceValueToInterfaceCompact(t *testing.T) {

	testCases := []Cadencetest{
		{"nil", nil, nil},
		{"None", cadence.NewOptional(nil), nil},
		{"Some(string)", cadence.NewOptional(CadenceString("foo")), "foo"},
		{"Some(int)", cadence.NewOptional(cadence.NewUInt64(42)), 42},
	}

	for _, tc := range testCases {
		//		t.Parallel()

		t.Run(tc.name, func(t *testing.T) {
			value := CadenceValueToInterfaceCompact(tc.value)
			assert.Equal(t, tc.result, value)
		})
	}
}

/*
		t.Run("Array", func(t *testing.T) {
			value := CadenceValueToJsonString(cadence.NewArray([]cadence.Value{CadenceString("foo"), CadenceString("bar")}))
			assert.Equal(t, `[
	    "foo",
	    "bar"
	]`, value)
		})

		t.Run("Dictionary", func(t *testing.T) {
			dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: CadenceString("foo"), Value: CadenceString("bar")}})
			value := CadenceValueToJsonString(dict)
			assert.Equal(t, `{
	    "foo": "bar"
	}`, value)
		})

		t.Run("Dictionary nested", func(t *testing.T) {
			subDict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: CadenceString("foo"), Value: CadenceString("bar")}})
			dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: CadenceString("foo"), Value: CadenceString("bar")}, {Key: CadenceString("subdict"), Value: subDict}})
			value := CadenceValueToJsonString(dict)
			assert.Equal(t, `{
	    "foo": "bar",
	    "subdict": {
	        "foo": "bar"
	    }
	}`, value)
		})

		t.Run("Dictionary", func(t *testing.T) {
			dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: cadence.NewUInt64(1), Value: cadence.NewUInt64(1)}})
			value := CadenceValueToJsonString(dict)
			assert.Equal(t, `{
	    "1": "1"
	}`, value)
		})

		t.Run("Struct", func(t *testing.T) {

			s := cadence.Struct{
				Fields: []cadence.Value{CadenceString("bar")},
				StructType: &cadence.StructType{
					Fields: []cadence.Field{{
						Identifier: "foo",
						Type:       cadence.StringType{},
					}},
				},
			}
			value := CadenceValueToJsonString(s)
			assert.Equal(t, `{
	    "foo": "bar"
	}`, value)
		})

		t.Run("Address", func(t *testing.T) {
			account := flow.HexToAddress("0x1")
			accountArg := cadence.BytesToAddress(account.Bytes())

			value := CadenceValueToJsonString(accountArg)
			assert.Equal(t, "\"0x0000000000000001\"", value)
		})

*/

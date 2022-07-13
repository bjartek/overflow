package overflow

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

func TestCadenceValueToJsonString(t *testing.T) {
	t.Parallel()
	t.Run("Empty value should be empty json object", func(t *testing.T) {
		value := CadenceValueToJsonString(nil)
		assert.Equal(t, "{}", value)
	})

	t.Run("Empty optional should be empty string", func(t *testing.T) {
		value := CadenceValueToJsonString(cadence.NewOptional(nil))
		assert.Equal(t, `""`, value)
	})

	t.Run("Emoji", func(t *testing.T) {
		value := CadenceValueToJsonString(cadence.NewOptional(CadenceString("😁")))
		assert.Equal(t, `"😁"`, value)
	})

	t.Run("Unwrap optional", func(t *testing.T) {
		value := CadenceValueToJsonString(cadence.NewOptional(CadenceString("foo")))
		assert.Equal(t, `"foo"`, value)
	})
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

	t.Run("Dictionary with numbers", func(t *testing.T) {
		fix64, err := cadence.NewFix64("1.12345678")
		assert.Nil(t, err)
		ufix64, err := cadence.NewUFix64("2.12345678")
		assert.Nil(t, err)

		dict := cadence.NewDictionary([]cadence.KeyValuePair{
			{Key: CadenceString("fix64"), Value: fix64},
			{Key: CadenceString("ufix64"), Value: ufix64},
		})
		value := CadenceValueToJsonString(dict)
		assert.Equal(t, `{
    "fix64": 1.12345678,
    "ufix64": 2.12345678
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

}

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
		// fixed point numbers
		fix64, err := cadence.NewFix64("1.12345678")
		assert.Nil(t, err)
		ufix64, err := cadence.NewUFix64("2.12345678")
		assert.Nil(t, err)

		// unsigned integers
		uint256Val := cadence.NewUInt256(256)
		uint128Val := cadence.NewUInt128(128)
		uint64Val := cadence.NewUInt64(64)
		uint32Val := cadence.NewUInt32(32)
		uint16Val := cadence.NewUInt16(16)
		uint8Val := cadence.NewUInt8(8)

		// integers
		int256Val := cadence.NewInt256(256)
		int128Val := cadence.NewInt256(128)
		int64Val := cadence.NewInt64(64)
		int32Val := cadence.NewInt32(32)
		int16Val := cadence.NewInt16(16)
		int8Val := cadence.NewInt8(16)

		boolVal := cadence.NewBool(false)

		dict := cadence.NewDictionary([]cadence.KeyValuePair{
			{Key: CadenceString("fix64"), Value: fix64},
			{Key: CadenceString("ufix64"), Value: ufix64},
			{Key: CadenceString("uint256"), Value: uint256Val},
			{Key: CadenceString("uint128"), Value: uint128Val},
			{Key: CadenceString("uint64"), Value: uint64Val},
			{Key: CadenceString("uint32"), Value: uint32Val},
			{Key: CadenceString("uint16"), Value: uint16Val},
			{Key: CadenceString("uint8"), Value: uint8Val},
			{Key: CadenceString("int256"), Value: int256Val},
			{Key: CadenceString("int128"), Value: int128Val},
			{Key: CadenceString("int64"), Value: int64Val},
			{Key: CadenceString("int32"), Value: int32Val},
			{Key: CadenceString("int16"), Value: int16Val},
			{Key: CadenceString("int8"), Value: int8Val},
			{Key: CadenceString("bool"), Value: boolVal},
		})
		value := CadenceValueToJsonString(dict)
		assert.Equal(t, `{
    "bool": false,
    "fix64": 1.12345678,
    "int128": 128,
    "int16": 16,
    "int256": 256,
    "int32": 32,
    "int64": 64,
    "int8": 16,
    "ufix64": 2.12345678,
    "uint128": 128,
    "uint16": 16,
    "uint256": 256,
    "uint32": 32,
    "uint64": 64,
    "uint8": 8
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
    "1": 1
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

package overflow

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/onflow/cadence"
	"github.com/onflow/cadence/runtime/common"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

type Cadencetest struct {
	want  autogold.Value
	input cadence.Value
}

func TestCadenceValueToInterface(t *testing.T) {

	foo := cadenceString("foo")
	bar := cadenceString("bar")
	emptyString := cadenceString("")

	emptyStrct := cadence.Struct{
		Fields: []cadence.Value{emptyString},
		StructType: &cadence.StructType{
			Fields: []cadence.Field{{
				Identifier: "foo",
				Type:       cadence.StringType{},
			}},
		},
	}

	address1 := flow.HexToAddress("f8d6e0586b0a20c7")
	caddress1, _ := common.BytesToAddress(address1.Bytes())
	structType := cadence.StructType{
		Location:            common.NewAddressLocation(nil, caddress1, ""),
		QualifiedIdentifier: "Contract.Bar",
		Fields: []cadence.Field{{
			Identifier: "foo",
			Type:       cadence.StringType{},
		}},
	}
	strct := cadence.Struct{
		Fields:     []cadence.Value{bar},
		StructType: &structType,
	}
	dict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: foo, Value: bar}})

	emoji := cadenceString("😁")
	emojiDict := cadence.NewDictionary([]cadence.KeyValuePair{{Key: emoji, Value: emoji}})

	cadenceAddress1 := cadence.BytesToAddress(address1.Bytes())

	structTypeValue := cadence.NewTypeValue(&structType)
	stringType := cadence.NewStringType()
	stringTypeValue := cadence.NewTypeValue(&stringType)
	ufix, _ := cadence.NewUFix64("42.0")
	path := cadence.Path{Domain: "storage", Identifier: "foo"}

	testCases := []Cadencetest{
		{autogold.Want("EmptyString", nil), cadenceString("")},
		{autogold.Want("nil", nil), nil},
		{autogold.Want("None", nil), cadence.NewOptional(nil)},
		{autogold.Want("Some(string)", "foo"), cadence.NewOptional(foo)},
		{autogold.Want("Some(uint64)", uint64(42)), cadence.NewOptional(cadence.NewUInt64(42))},
		{autogold.Want("uint64", uint64(42)), cadence.NewUInt64(42)},
		{autogold.Want("ufix64", float64(42.0)), ufix},
		{autogold.Want("uint32", uint32(42)), cadence.NewUInt32(42)},
		{autogold.Want("int", 42), cadence.NewInt(42)},
		{autogold.Want("string array", []interface{}{"foo", "bar"}), cadence.NewArray([]cadence.Value{foo, bar})},
		{autogold.Want("empty array", nil), cadence.NewArray([]cadence.Value{emptyString})},
		{autogold.Want("string array ignore empty", []interface{}{"foo", "bar"}), cadence.NewArray([]cadence.Value{foo, emptyString, bar})},
		{autogold.Want("dictionary", map[string]interface{}{"foo": "bar"}), dict},
		{autogold.Want("dictionary_ignore_empty_value", nil), cadence.NewDictionary([]cadence.KeyValuePair{{Key: foo, Value: emptyString}})},
		{autogold.Want("dictionary_with_subdict", map[string]interface{}{"bar": map[string]interface{}{"foo": "bar"}}), cadence.NewDictionary([]cadence.KeyValuePair{{Key: bar, Value: dict}})},
		{autogold.Want("struct", map[string]interface{}{"foo": "bar"}), strct},
		{autogold.Want("empty struct", nil), emptyStrct},
		{autogold.Want("address", "0xf8d6e0586b0a20c7"), cadenceAddress1},
		{autogold.Want("string type", "String"), stringTypeValue},
		{autogold.Want("struct type", "A.f8d6e0586b0a20c7.Contract.Bar"), structTypeValue},
		{autogold.Want("Emoji", "😁"), emoji},
		{autogold.Want("EmojiDict", map[string]interface{}{"😁": "😁"}), emojiDict},
		{autogold.Want("StoragePath", "/storage/foo"), path},
	}

	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			value := CadenceValueToInterface(tc.input)
			tc.want.Equal(t, value)
		})
	}
}

func TestCadenceValueToJson(t *testing.T) {
	result, err := CadenceValueToJsonString(cadence.String(""))
	assert.NoError(t, err)
	assert.Equal(t, "", result)

}

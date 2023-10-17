package overflow

import (
	"encoding/json"
	"fmt"
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

	stringType := cadence.NewStringType()
	cadenceEvent := cadence.NewEvent([]cadence.Value{foo}).WithType(&cadence.EventType{
		QualifiedIdentifier: "TestEvent",
		Fields: []cadence.Field{{
			Type:       cadence.StringType{},
			Identifier: "foo",
		}},
	},
	)

	structTypeValue := cadence.NewTypeValue(&structType)
	stringTypeValue := cadence.NewTypeValue(&stringType)
	ufix, _ := cadence.NewUFix64("42.0")
	path := cadence.Path{Domain: common.PathDomainStorage, Identifier: "foo"}

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
		{autogold.Want("Event", map[string]interface{}{"foo": "foo"}), cadenceEvent},
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

func TestParseInputValue(t *testing.T) {
	foo := "foo"

	var strPointer *string = nil
	values := []interface{}{
		"foo",
		uint64(42),
		map[string]uint64{"foo": uint64(42)},
		[]uint64{42, 69},
		[2]string{"foo", "bar"},
		&foo,
		strPointer,
	}

	for idx, value := range values {
		t.Run(fmt.Sprintf("parse input %d", idx), func(t *testing.T) {
			cv, err := InputToCadence(value, func(string) (string, error) {
				return "", nil
			})
			assert.NoError(t, err)
			v := CadenceValueToInterface(cv)

			vj, err := json.Marshal(v)
			assert.NoError(t, err)

			cvj, err := json.Marshal(value)
			assert.NoError(t, err)

			assert.Equal(t, string(cvj), string(vj))
		})
	}
}

func TestMarshalCadenceStruct(t *testing.T) {
	val, err := InputToCadence(Foo{Bar: "foo"}, func(string) (string, error) {
		return "A.123.Foo.Bar", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "A.123.Foo.Bar", val.Type().ID())
	jsonVal, err := CadenceValueToJsonString(val)
	assert.NoError(t, err)
	assert.JSONEq(t, `{ "bar": "foo" }`, jsonVal)
}

func TestMarshalCadenceStructWithStructTag(t *testing.T) {
	val, err := InputToCadence(Foo{Bar: "foo"}, func(string) (string, error) {
		return "A.123.Foo.Baz", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "A.123.Foo.Baz", val.Type().ID())
	jsonVal, err := CadenceValueToJsonString(val)
	assert.NoError(t, err)
	assert.JSONEq(t, `{ "bar": "foo" }`, jsonVal)
}

func TestPrimitiveInputToCadence(t *testing.T) {
	tests := []struct {
		name  string
		value interface{}
	}{
		{name: "int", value: 1},
		{name: "int8", value: int8(8)},
		{name: "int16", value: int16(16)},
		{name: "int32", value: int32(32)},
		{name: "int64", value: int64(64)},
		{name: "uint8", value: uint8(8)},
		{name: "uint16", value: uint16(16)},
		{name: "uint32", value: uint32(32)},
		{name: "true", value: true},
		{name: "false", value: false},
	}

	resolver := func(string) (string, error) {
		return "", nil
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cadenceValue, err := InputToCadence(test.value, resolver)
			assert.NoError(t, err)
			result2 := CadenceValueToInterface(cadenceValue)
			assert.Equal(t, test.value, result2)
		})
	}
}

// in Debug.cdc
type Foo struct {
	Bar string
}

type Debug_FooListBar struct {
	Bar string
	Foo []Debug_Foo2
}

type Debug_FooBar struct {
	Bar string
	Foo Debug_Foo
}

type Debug_Foo_Skip struct {
	Bar  string
	Skip string `cadence:"-"`
}

type Debug_Foo2 struct {
	Bar string `cadence:"bar,cadenceAddress"`
}

type Debug_Foo struct {
	Bar string
}

// in Foo.Bar.Baz
type Baz struct {
	Something string `json:"bar"`
}

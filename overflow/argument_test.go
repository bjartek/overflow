package overflow

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk"
	"github.com/stretchr/testify/assert"
)

/*
 Tests must be in the same folder as flow.json with contracts and transactions/scripts in subdirectories in order for the path resolver to work correctly
*/
func TestArguments(t *testing.T) {
	g := NewTestingEmulator().Start()
	t.Parallel()

	t.Run("Argument test", func(t *testing.T) {

		fix, _ := cadence.NewFix64("-1.0")
		ufix, _ := cadence.NewUFix64("1.0")
		dateFix, _ := cadence.NewUFix64("1627560000.00000000")

		foo, _ := cadence.NewString("foo")
		bar, _ := cadence.NewString("bar")

		address1 := flow.HexToAddress("f8d6e0586b0a20c7")
		address2 := flow.HexToAddress("01cf0e2f2f715450")
		cadenceAddress1 := cadence.BytesToAddress(address1.Bytes())
		cadenceAddress2 := cadence.BytesToAddress(address2.Bytes())

		stringValues := []cadence.Value{foo, bar}
		addressValues := []cadence.Value{cadenceAddress1, cadenceAddress2}

		builder := g.Arguments().Boolean(true).
			Bytes([]byte{1}).
			Fix64("-1.0").
			UFix64(1.0).
			DateStringAsUnixTimestamp("July 29, 2021 08:00:00 AM", "America/New_York").
			StringArray("foo", "bar").
			RawAddressArray("f8d6e0586b0a20c7", "01cf0e2f2f715450")

		assert.Contains(t, builder.Arguments, cadence.NewBool(true))
		assert.Contains(t, builder.Arguments, cadence.NewBytes([]byte{1}))
		assert.Contains(t, builder.Arguments, fix)
		assert.Contains(t, builder.Arguments, ufix)
		assert.Contains(t, builder.Arguments, dateFix)
		assert.Contains(t, builder.Arguments, cadence.NewArray(stringValues))
		assert.Contains(t, builder.Arguments, cadence.NewArray(addressValues))
	})

	t.Run("Word argument test", func(t *testing.T) {
		builder := g.Arguments().
			Word8(8).
			Word16(16).
			Word32(32).
			Word64(64)

		assert.Contains(t, builder.Arguments, cadence.NewWord8(8))
		assert.Contains(t, builder.Arguments, cadence.NewWord16(16))
		assert.Contains(t, builder.Arguments, cadence.NewWord32(32))
		assert.Contains(t, builder.Arguments, cadence.NewWord64(64))
	})

	t.Run("UInt argument test", func(t *testing.T) {
		builder := g.Arguments().
			UInt(1).
			UInt8(8).
			UInt16(16).
			UInt32(32).
			UInt64(64).
			UInt128(128).
			UInt256(256)

		assert.Contains(t, builder.Arguments, cadence.NewUInt(1))
		assert.Contains(t, builder.Arguments, cadence.NewUInt8(8))
		assert.Contains(t, builder.Arguments, cadence.NewUInt16(16))
		assert.Contains(t, builder.Arguments, cadence.NewUInt32(32))
		assert.Contains(t, builder.Arguments, cadence.NewUInt64(64))
		assert.Contains(t, builder.Arguments, cadence.NewUInt128(128))
		assert.Contains(t, builder.Arguments, cadence.NewUInt256(256))
	})

	t.Run("Int argument test", func(t *testing.T) {
		builder := g.Arguments().
			Int(1).
			Int8(-8).
			Int16(16).
			Int32(32).
			Int64(64).
			Int128(128).
			Int256(256)

		assert.Contains(t, builder.Arguments, cadence.NewInt(1))
		assert.Contains(t, builder.Arguments, cadence.NewInt8(-8))
		assert.Contains(t, builder.Arguments, cadence.NewInt16(16))
		assert.Contains(t, builder.Arguments, cadence.NewInt32(32))
		assert.Contains(t, builder.Arguments, cadence.NewInt64(64))
		assert.Contains(t, builder.Arguments, cadence.NewInt128(128))
		assert.Contains(t, builder.Arguments, cadence.NewInt256(256))
	})

	t.Run("Raw address", func(t *testing.T) {
		builder := g.Arguments().RawAddress("0x01cf0e2f2f715450")
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("Address", func(t *testing.T) {
		builder := g.Arguments().Address("first")
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("Fix64 error", func(t *testing.T) {
		builder := g.Arguments().Fix64("first")
		assert.ErrorContains(t, builder.Error, "missing decimal point")
	})

	t.Run("Fix64 error build", func(t *testing.T) {
		assert.Panics(t, func() {
			g.Arguments().Fix64("first").Build()
		})
	})

	t.Run("UFix64 error", func(t *testing.T) {
		builder := g.Arguments().UFix64(-1.0)
		assert.ErrorContains(t, builder.Error, "invalid negative integer part")
	})

	t.Run("UFix64Array error", func(t *testing.T) {
		builder := g.Arguments().UFix64Array(-1.0)
		assert.ErrorContains(t, builder.Error, "invalid negative integer part")
	})

	t.Run("DateTime wrong locale", func(t *testing.T) {
		builder := g.Arguments().DateStringAsUnixTimestamp("asd", "asd")
		assert.ErrorContains(t, builder.Error, "unknown time zone as")
	})

	t.Run("DateTime wrong string", func(t *testing.T) {
		builder := g.Arguments().DateStringAsUnixTimestamp("asd", "Europe/Oslo")
		assert.ErrorContains(t, builder.Error, "Could not find format for \"asd\"")
	})

	t.Run("Paths", func(t *testing.T) {
		builder := g.Arguments().PublicPath("foo").StoragePath("foo").PrivatePath("foo")
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("StringMap", func(t *testing.T) {
		builder := g.Arguments().StringMap(map[string]string{"foo": "bar"})
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("ScalarMap", func(t *testing.T) {
		builder := g.Arguments().ScalarMap(map[string]string{"foo": "1.0"})
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("ScalarMapError", func(t *testing.T) {
		builder := g.Arguments().ScalarMap(map[string]string{"foo": "foo"})
		assert.ErrorContains(t, builder.Error, "missing decimal point")
	})

	t.Run("StringArray", func(t *testing.T) {
		builder := g.Arguments().StringArray("foo", "bar")
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("StringMapArray", func(t *testing.T) {
		builder := g.Arguments().StringMapArray(map[string]string{"Sith": "Darth Vader"}, map[string]string{"Jedi": "Luke"})
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("ScalarMapArray", func(t *testing.T) {
		builder := g.Arguments().ScalarMapArray(map[string]string{"Sith": "2.0"}, map[string]string{"Jedi": "1000.0"})
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("ScalarMapArray error", func(t *testing.T) {
		builder := g.Arguments().ScalarMapArray(map[string]string{"Sith": "2.0"}, map[string]string{"Jedi": "asd"})
		assert.ErrorContains(t, builder.Error, "missing decimal point")
	})

	t.Run("AccountArray", func(t *testing.T) {
		builder := g.Arguments().AccountArray("first", "second")
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("Uint64Array", func(t *testing.T) {
		builder := g.Arguments().UInt64Array(1, 2, 3)
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("Uint8Array", func(t *testing.T) {
		builder := g.Arguments().UInt8Array(1, 2, 3)
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("UFix64Array", func(t *testing.T) {
		builder := g.Arguments().UFix64Array(1.0, 2.0, 3.0)
		autogold.Equal(t, builder.Arguments)
	})

	t.Run("Argument string array", func(t *testing.T) {
		builder := g.Transaction(`
transaction(names: [String]) {

}
`).NamedArguments(map[string]string{
			"names": `["Bjarte", "Karlsen"]`,
		})

		arrayValues := []cadence.Value{
			CadenceString("Bjarte"),
			CadenceString("Karlsen"),
		}
		assert.Contains(t, builder.Arguments, cadence.NewArray(arrayValues))

	})

	t.Run("Argument ufix64 array", func(t *testing.T) {
		builder := g.Transaction(`
transaction(names: [UFix64]) {

}
`).NamedArguments(map[string]string{
			"names": `[10.0, 20.0]`,
		})

		fix, _ := cadence.NewUFix64("10.0")
		fix1, _ := cadence.NewUFix64("20.0")
		arrayValues := []cadence.Value{fix, fix1}
		assert.Contains(t, builder.Arguments, cadence.NewArray(arrayValues))

	})

	/*
			how is byte array represented in cadence again?
			t.Run("Argument byte array", func(t *testing.T) {
				bytes := []byte("test")

				builder := g.Transaction(`
		transaction(arg: [UInt8]) {

		}
		`).NamedArguments(map[string]string{
					"arg": string(bytes),
				})
				assert.NoError(t, builder.Error)

				fmt.Printf("%+v", builder.Arguments)
				expected := cadence.NewBytes(bytes)
				assert.Contains(t, builder.Arguments, expected)

			})
	*/
}

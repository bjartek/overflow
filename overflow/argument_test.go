package overflow

import (
	"testing"

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
			AddressArray("f8d6e0586b0a20c7", "01cf0e2f2f715450")

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
}

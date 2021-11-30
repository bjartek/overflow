	package overflow

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/stretchr/testify/assert"
)

func TestSetupFails(t *testing.T) {

	g := NewGoWithTheFlow([]string{"../examples/flow.json"}, "emulator", true, output.NoneLog)
	_, err := g.CreateAccountsE("foobar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find account with name foobar")

}
func TestScriptArguments(t *testing.T) {
	g := NewGoWithTheFlow([]string{"../examples/flow.json"}, "emulator", true, output.NoneLog)
	t.Parallel()

	t.Run("Argument test", func(t *testing.T) {

		fix, _ := cadence.NewFix64("-1.0")
		ufix, _ := cadence.NewUFix64("1.0")

		dateFix, _ := cadence.NewUFix64("1627560000.00000000")

		builder := g.Script("").BooleanArgument(true).
			BytesArgument([]byte{1}).
			Fix64Argument("-1.0").
			UFix64Argument("1.0").
			StringArgument("test").
			DateStringAsUnixTimestamp("July 29, 2021 08:00:00 AM", "America/New_York")

		string, _ := cadence.NewString("test")
		assert.Contains(t, builder.Arguments, cadence.NewBool(true))
		assert.Contains(t, builder.Arguments, cadence.NewBytes([]byte{1}))
		assert.Contains(t, builder.Arguments, fix)
		assert.Contains(t, builder.Arguments, ufix)
		assert.Contains(t, builder.Arguments, string)
		assert.Contains(t, builder.Arguments, dateFix)

	})

	t.Run("Word argument test", func(t *testing.T) {
		builder := g.Script("").
			Word8Argument(8).
			Word16Argument(16).
			Word32Argument(32).
			Word64Argument(64)

		assert.Contains(t, builder.Arguments, cadence.NewWord8(8))
		assert.Contains(t, builder.Arguments, cadence.NewWord16(16))
		assert.Contains(t, builder.Arguments, cadence.NewWord32(32))
		assert.Contains(t, builder.Arguments, cadence.NewWord64(64))
	})

	t.Run("UInt argument test", func(t *testing.T) {
		builder := g.Script("").
			UIntArgument(1).
			UInt8Argument(8).
			UInt16Argument(16).
			UInt32Argument(32).
			UInt64Argument(64).
			UInt128Argument(128).
			UInt256Argument(256)

		assert.Contains(t, builder.Arguments, cadence.NewUInt(1))
		assert.Contains(t, builder.Arguments, cadence.NewUInt8(8))
		assert.Contains(t, builder.Arguments, cadence.NewUInt16(16))
		assert.Contains(t, builder.Arguments, cadence.NewUInt32(32))
		assert.Contains(t, builder.Arguments, cadence.NewUInt64(64))
		assert.Contains(t, builder.Arguments, cadence.NewUInt128(128))
		assert.Contains(t, builder.Arguments, cadence.NewUInt256(256))
	})

	t.Run("Int argument test", func(t *testing.T) {
		builder := g.Script("").
			IntArgument(1).
			Int8Argument(-8).
			Int16Argument(16).
			Int32Argument(32).
			Int64Argument(64).
			Int128Argument(128).
			Int256Argument(256)

		assert.Contains(t, builder.Arguments, cadence.NewInt(1))
		assert.Contains(t, builder.Arguments, cadence.NewInt8(-8))
		assert.Contains(t, builder.Arguments, cadence.NewInt16(16))
		assert.Contains(t, builder.Arguments, cadence.NewInt32(32))
		assert.Contains(t, builder.Arguments, cadence.NewInt64(64))
		assert.Contains(t, builder.Arguments, cadence.NewInt128(128))
		assert.Contains(t, builder.Arguments, cadence.NewInt256(256))
	})
}

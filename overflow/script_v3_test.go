package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScript(t *testing.T) {
	o, _ := OverflowTesting()

	t.Run("Run simple script interface", func(t *testing.T) {
		res, err := o.Script("test", Arg("account", "first")).GetAsInterface()
		assert.NoError(t, err)
		assert.Equal(t, "0x01cf0e2f2f715450", res)
	})

	t.Run("Run simple script json", func(t *testing.T) {
		res, err := o.Script("test", Arg("account", "first")).GetAsJson()
		assert.NoError(t, err)
		assert.Equal(t, `"0x01cf0e2f2f715450"`, res)
	})

	t.Run("Run simple script marshal", func(t *testing.T) {
		var res string
		err := o.Script("test", Arg("account", "first")).MarhalAs(&res)
		assert.NoError(t, err)
		assert.Equal(t, "0x01cf0e2f2f715450", res)
	})

	t.Run("compose a script", func(t *testing.T) {
		accountScript := o.ScriptFN(Arg("account", "first"))
		res := accountScript("test")
		assert.NoError(t, res.Err)
	})

	t.Run("create script with name", func(t *testing.T) {
		testScript := o.ScriptFileNameFN("test")
		res := testScript(Arg("account", "first"))
		assert.NoError(t, res.Err)
	})

	t.Run("Run script inline", func(t *testing.T) {
		res, err := o.Script(`
pub fun main(): String {
	return "foo"
}
`).GetAsJson()
		assert.NoError(t, err)
		assert.Equal(t, `"foo"`, res)
	})

	t.Run("Run script return type", func(t *testing.T) {
		res, err := o.Script("type").GetAsInterface()
		assert.NoError(t, err)
		assert.Equal(t, `A.0ae53cb6e3f42a79.FlowToken.Vault`, res)
	})

	t.Run("Run script with ufix64 array", func(t *testing.T) {
		res, err := o.Script(`
pub fun main(input: [UFix64]): [UFix64] {
	return input
}

`, Arg("input", `[10.1, 20.2]`)).GetAsJson()
		assert.NoError(t, err)
		assert.JSONEq(t, `[10.1, 20.2]`, res)
	})

	t.Run("Run script with fix64 array", func(t *testing.T) {
		res, err := o.Script(`
pub fun main(input: [Fix64]): [Fix64] {
	return input
}

`, Arg("input", `[10.1, -20.2]`)).GetAsJson()
		assert.NoError(t, err)

		assert.JSONEq(t, `[10.1, -20.2]`, res)
	})

	t.Run("Run script with uint64 array", func(t *testing.T) {
		res, err := o.Script(`
pub fun main(input: [UInt64]): [UInt64] {
	return input
}

`, Arg("input", `[10, 20]`)).GetAsJson()

		assert.NoError(t, err)
		assert.JSONEq(t, `[10, 20]`, res)
	})

}

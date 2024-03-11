package overflow

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScript(t *testing.T) {
	o, err := OverflowTesting()
	require.NoError(t, err)
	require.NotNil(t, o)

	t.Run("Run simple script interface", func(t *testing.T) {
		res, err := o.Script("test", WithArg("account", "first")).GetAsInterface()
		assert.NoError(t, err)
		assert.Equal(t, "0x179b6b1cb6755e31", res)
	})

	t.Run("Run simple script json", func(t *testing.T) {
		res, err := o.Script("test", WithArg("account", "first")).GetAsJson()
		assert.NoError(t, err)
		assert.Equal(t, `"0x179b6b1cb6755e31"`, res)
	})

	t.Run("Run simple script marshal", func(t *testing.T) {
		var res string
		err := o.Script("test", WithArg("account", "first")).MarshalAs(&res)
		assert.NoError(t, err)
		assert.Equal(t, "0x179b6b1cb6755e31", res)
	})

	t.Run("Run simple script marshal with underlying error", func(t *testing.T) {
		var res string
		err := o.Script("test2", WithArg("account", "first")).MarshalAs(&res)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "Could not read interaction file from path")
	})

	t.Run("compose a script", func(t *testing.T) {
		accountScript := o.ScriptFN(WithArg("account", "first"))
		res := accountScript("test")
		assert.NoError(t, res.Err)
	})

	t.Run("create script with name", func(t *testing.T) {
		testScript := o.ScriptFileNameFN("test")
		res := testScript(WithArg("account", "first"))
		assert.NoError(t, res.Err)
	})

	t.Run("Run script inline", func(t *testing.T) {
		res, err := o.Script(`
access(all) fun main(): String {
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

	t.Run("Run script with string array", func(t *testing.T) {
		input := []string{"test", "foo"}
		res, err := o.Script(`
access(all) fun main(input: [String]): [String] {
	return input
}

`, WithArg("input", input)).GetAsJson()
		assert.NoError(t, err)
		assert.JSONEq(t, `["test", "foo"]`, res)
	})

	t.Run("Run script with string map", func(t *testing.T) {
		input := `{"test": "foo", "test2": "bar"}`
		res, err := o.Script(`
access(all) fun main(input: {String : String}): {String: String} {
	return input
}

`, WithArg("input", input)).GetAsJson()
		assert.NoError(t, err)
		assert.JSONEq(t, `{"test": "foo", "test2":"bar"}`, res)
	})

	t.Run("Run script with string map from native input", func(t *testing.T) {
		input := map[string]string{
			"test":  "foo",
			"test2": "bar",
		}
		res, err := o.Script(`
access(all) fun main(input: {String : String}): {String: String} {
	return input
}

`, WithArg("input", input)).GetAsJson()
		assert.NoError(t, err)
		assert.JSONEq(t, `{"test": "foo", "test2":"bar"}`, res)
	})

	t.Run("Run script with string:float64 map from native input", func(t *testing.T) {
		input := map[string]float64{
			"test":  1.0,
			"test2": 2.0,
		}
		res, err := o.Script(`
access(all) fun main(input: {String : UFix64}): {String: UFix64} {
	return input
}

`, WithArg("input", input)).GetAsJson()
		assert.NoError(t, err)
		assert.JSONEq(t, `{"test": 1.0, "test2":2.0}`, res)
	})

	t.Run("Run script with string:uint64 map from native input", func(t *testing.T) {
		input := map[string]uint64{
			"test":  1,
			"test2": 2,
		}
		res, err := o.Script(`
access(all) fun main(input: {String : UInt64}): {String: UInt64} {
	return input
}

`, WithArg("input", input)).GetAsJson()
		assert.NoError(t, err)
		assert.JSONEq(t, `{"test": 1, "test2":2}`, res)
	})

	t.Run("Run script with ufix64 array as string", func(t *testing.T) {
		res, err := o.Script(`
access(all) fun main(input: [UFix64]): [UFix64] {
	return input
}

`, WithArg("input", `[10.1, 20.2]`)).GetAsJson()
		assert.NoError(t, err)
		assert.JSONEq(t, `[10.1, 20.2]`, res)
	})

	t.Run("Run script with ufix64 array", func(t *testing.T) {
		res, err := o.Script(`
access(all) fun main(input: [UFix64]): [UFix64] {
	return input
}

`, WithArg("input", []float64{10.1, 20.2})).GetAsJson()
		assert.NoError(t, err)
		assert.JSONEq(t, `[10.1, 20.2]`, res)
	})

	t.Run("Run script with fix64 array", func(t *testing.T) {
		res, err := o.Script(`
access(all) fun main(input: [Fix64]): [Fix64] {
	return input
}

`, WithArg("input", `[10.1, -20.2]`)).GetAsJson()
		assert.NoError(t, err)

		assert.JSONEq(t, `[10.1, -20.2]`, res)
	})

	t.Run("Run script with uint64 array as string", func(t *testing.T) {
		res, err := o.Script(`
access(all) fun main(input: [UInt64]): [UInt64] {
	return input
}

`, WithArg("input", `[10, 20]`)).GetAsJson()

		assert.NoError(t, err)
		assert.JSONEq(t, `[10, 20]`, res)
	})

	t.Run("Run script with uint64 array", func(t *testing.T) {
		res, err := o.Script(`
access(all) fun main(input: [UInt64]): [UInt64] {
	return input
}

`, WithArg("input", []uint64{10, 20})).GetAsJson()

		assert.NoError(t, err)
		assert.JSONEq(t, `[10, 20]`, res)
	})

	t.Run("Run script with optional Address some", func(t *testing.T) {
		res, err := o.Script(`
access(all) fun main(input: Address?): Address? {
	return input
}

`, WithArg("input", "0x01cf0e2f2f715450")).GetAsInterface()

		assert.NoError(t, err)
		assert.Equal(t, `0x01cf0e2f2f715450`, res)
	})

	t.Run("Run script with optional Address empty", func(t *testing.T) {
		res, err := o.Script(`
access(all) fun main(input: Address?): Address? {
	return input
}

`, WithArg("input", nil)).GetAsInterface()

		assert.NoError(t, err)
		assert.Equal(t, nil, res)
	})

	t.Run("Run script at previous height", func(t *testing.T) {
		block, err := o.GetLatestBlock(context.Background())
		require.NoError(t, err)
		res, err := o.Script("test", WithArg("account", "first"), WithExecuteScriptAtBlockHeight(block.Height-1)).GetAsInterface()
		assert.NoError(t, err)
		assert.Equal(t, "0x179b6b1cb6755e31", res)
	})

	t.Run("Run script at block", func(t *testing.T) {
		block, err := o.GetLatestBlock(context.Background())
		require.NoError(t, err)
		block, err = o.GetBlockAtHeight(context.Background(), block.Height-1)
		assert.NoError(t, err)
		res, err := o.Script("test", WithArg("account", "first"), WithExecuteScriptAtBlockIdentifier(block.ID)).GetAsInterface()
		assert.NoError(t, err)
		assert.Equal(t, "0x179b6b1cb6755e31", res)
	})
	t.Run("Run script at block hex", func(t *testing.T) {
		block, err := o.GetLatestBlock(context.Background())
		require.NoError(t, err)
		block, err = o.GetBlockAtHeight(context.Background(), block.Height-1)
		assert.NoError(t, err)
		res, err := o.Script("test", WithArg("account", "first"), WithExecuteScriptAtBlockIdHex(block.ID.Hex())).GetAsInterface()
		assert.NoError(t, err)
		assert.Equal(t, "0x179b6b1cb6755e31", res)
	})
}

package overflow

import (
	"fmt"
	"testing"

	"github.com/hexops/autogold"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig(t *testing.T) {
	g, err := OverflowTesting()
	require.NoError(t, err)
	require.NotNil(t, g)

	t.Run("parse", func(t *testing.T) {
		result, err := g.ParseAll()
		assert.NoError(t, err)

		autogold.Equal(t, result)
	})

	t.Run("parse and merge", func(t *testing.T) {
		result, err := g.ParseAll()
		assert.NoError(t, err)

		merged := result.MergeSpecAndCode()
		autogold.Equal(t, merged)
	})

	t.Run("parse and filter", func(t *testing.T) {
		result, err := g.ParseAllWithConfig(true, []string{"arguments"}, []string{"test"})
		assert.NoError(t, err)
		autogold.Equal(t, result)
	})

	t.Run("parse and merge strip network prefix scripts", func(t *testing.T) {
		result, err := g.ParseAll()
		assert.NoError(t, err)

		merged := result.MergeSpecAndCode()
		emulator := merged.Networks["emulator"]
		_, ok := emulator.Scripts["Foo"]
		assert.True(t, ok, litter.Sdump(emulator.Scripts))
		mainnet := merged.Networks["mainnet"]
		_, mainnetOk := mainnet.Scripts["Foo"]
		assert.True(t, mainnetOk, litter.Sdump(mainnet.Scripts))

	})

	t.Run("parse and merge strip network prefix transaction", func(t *testing.T) {
		result, err := g.ParseAll()
		assert.NoError(t, err)
		merged := result.MergeSpecAndCode()
		emulator := merged.Networks["emulator"]
		_, ok := emulator.Transactions["Foo"]
		assert.True(t, ok, litter.Sdump(emulator.Transactions))
		mainnet := merged.Networks["mainnet"]
		_, mainnetOk := mainnet.Transactions["Foo"]
		assert.True(t, mainnetOk, litter.Sdump(mainnet.Transactions))

	})

	tx := `import FungibleToken from %s

transaction() {
  // This is a %s transaction
}`

	script := `import FungibleToken from %s
// This is a %s script

access(all) fun main(account: Address): String {
    return getAccount(account).address.toString()
}`

	result, err := g.ParseAll()
	assert.NoError(t, err)
	merged := result.MergeSpecAndCode()

	type ParseIntegrationTestCase struct {
		Network         string
		ScriptName      string
		ScriptType      string
		ContractAddress string
		Expected        string
	}

	tcs := []ParseIntegrationTestCase{
		{
			Network:         "mainnet",
			ScriptType:      "Transaction",
			ScriptName:      "aTransaction",
			ContractAddress: "0xf233dcee88fe0abe",
			Expected:        fmt.Sprintf(tx, "0xf233dcee88fe0abe", "mainnet specific"),
		},
		{
			Network:         "mainnet",
			ScriptType:      "Transaction",
			ScriptName:      "zTransaction",
			ContractAddress: "0xf233dcee88fe0abe",
			Expected:        fmt.Sprintf(tx, "0xf233dcee88fe0abe", "mainnet specific"),
		},
		{
			Network:         "emulator",
			ScriptType:      "Transaction",
			ScriptName:      "aTransaction",
			ContractAddress: "0xee82856bf20e2aa6",
			Expected:        fmt.Sprintf(tx, "0xee82856bf20e2aa6", "generic"),
		},
		{
			Network:         "emulator",
			ScriptType:      "Transaction",
			ScriptName:      "zTransaction",
			ContractAddress: "0xf233dcee88fe0abe",
			Expected:        fmt.Sprintf(tx, "0xee82856bf20e2aa6", "generic"),
		},
		{
			Network:         "mainnet",
			ScriptType:      "Script",
			ScriptName:      "aScript",
			ContractAddress: "0xf233dcee88fe0abe",
			Expected:        fmt.Sprintf(script, "0xf233dcee88fe0abe", "mainnet specific"),
		},
		{
			Network:         "mainnet",
			ScriptType:      "Script",
			ScriptName:      "zScript",
			ContractAddress: "0xf233dcee88fe0abe",
			Expected:        fmt.Sprintf(script, "0xf233dcee88fe0abe", "mainnet specific"),
		},
		{
			Network:         "emulator",
			ScriptType:      "Script",
			ScriptName:      "aScript",
			ContractAddress: "0xee82856bf20e2aa6",
			Expected:        fmt.Sprintf(script, "0xee82856bf20e2aa6", "generic"),
		},
		{
			Network:         "emulator",
			ScriptType:      "Script",
			ScriptName:      "zScript",
			ContractAddress: "0xf233dcee88fe0abe",
			Expected:        fmt.Sprintf(script, "0xee82856bf20e2aa6", "generic"),
		},
	}

	for _, tc := range tcs {
		t.Run(fmt.Sprintf("parse and overwrite %s with %s network prefix script : %s", tc.ScriptType, tc.Network, tc.ScriptName), func(t *testing.T) {
			network := merged.Networks[tc.Network]
			if tc.ScriptType == "Script" {
				script, ok := network.Scripts[tc.ScriptName]
				assert.True(t, ok, litter.Sdump(network.Scripts))
				assert.Equal(t, tc.Expected, script.Code)
				return
			}
			tx, ok := network.Transactions[tc.ScriptName]
			assert.True(t, ok, litter.Sdump(network.Transactions))
			assert.Equal(t, tc.Expected, tx.Code)
		})
	}

}

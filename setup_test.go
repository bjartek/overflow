package overflow

import (
	"testing"

	"github.com/onflow/flowkit/v2/output"
	"github.com/stretchr/testify/assert"
)

func TestOverflowv3(t *testing.T) {
	t.Run("WithNetworkEmbedded", func(t *testing.T) {
		b := Apply(WithNetwork("embedded"))
		assert.Equal(t, "emulator", b.Network)
		assert.True(t, b.DeployContracts)
		assert.True(t, b.InitializeAccounts)
		assert.True(t, b.InMemory)
		assert.Equal(t, output.NoneLog, b.LogLevel)
	})

	t.Run("WithNetworkTesting", func(t *testing.T) {
		b := Apply(WithNetwork("testing"))
		assert.Equal(t, "emulator", b.Network)
		assert.True(t, b.DeployContracts)
		assert.True(t, b.InitializeAccounts)
		assert.True(t, b.InMemory)
		assert.Equal(t, output.NoneLog, b.LogLevel)
	})

	t.Run("WithNetworkEmulator", func(t *testing.T) {
		b := Apply(WithNetwork("emulator"))
		assert.Equal(t, "emulator", b.Network)
		assert.True(t, b.DeployContracts)
		assert.True(t, b.InitializeAccounts)
		assert.False(t, b.InMemory)
	})

	t.Run("WithNetworkTestnet", func(t *testing.T) {
		b := Apply(WithNetwork("testnet"))
		assert.Equal(t, "testnet", b.Network)
		assert.False(t, b.DeployContracts)
		assert.False(t, b.InitializeAccounts)
		assert.False(t, b.InMemory)
	})

	t.Run("WithNetworkMainnet", func(t *testing.T) {
		b := Apply(WithNetwork("mainnet"))
		assert.Equal(t, "mainnet", b.Network)
		assert.False(t, b.DeployContracts)
		assert.False(t, b.InitializeAccounts)
		assert.False(t, b.InMemory)
	})

	t.Run("WithInMemory", func(t *testing.T) {
		b := Apply()
		assert.True(t, b.InMemory)
		assert.True(t, b.InitializeAccounts)
		assert.True(t, b.DeployContracts)
		assert.Equal(t, "emulator", b.Network)
	})

	t.Run("WithExistingEmulator", func(t *testing.T) {
		b := Apply(WithExistingEmulator())
		assert.False(t, b.InitializeAccounts)
		assert.False(t, b.DeployContracts)
	})

	t.Run("DoNotPrependNetworkToAccountNames", func(t *testing.T) {
		b := Apply(WithNoPrefixToAccountNames())
		assert.False(t, b.PrependNetworkName)
	})

	t.Run("WithServiceAccountSuffix", func(t *testing.T) {
		b := Apply(WithServiceAccountSuffix("foo"))
		assert.Equal(t, "foo", b.ServiceSuffix)
	})

	t.Run("WithBasePath", func(t *testing.T) {
		b := Apply(WithBasePath("../"))
		assert.Equal(t, "../", b.Path)
	})

	t.Run("WithNoLog", func(t *testing.T) {
		b := Apply(WithLogNone())
		assert.Equal(t, output.NoneLog, b.LogLevel)
	})

	t.Run("WithGas", func(t *testing.T) {
		b := Apply(WithGas(42))
		assert.Equal(t, 42, b.GasLimit)
	})

	t.Run("WithFlowConfig", func(t *testing.T) {
		b := Apply(WithFlowConfig("foo.json", "bar.json"))
		assert.Equal(t, []string{"foo.json", "bar.json"}, b.ConfigFiles)
	})

	t.Run("WithScriptFolderName", func(t *testing.T) {
		b := Apply(WithScriptFolderName("script"))
		assert.Equal(t, "script", b.ScriptFolderName)
	})

	t.Run("WithGlobalPrintOptions", func(t *testing.T) {
		b := Apply(WithGlobalPrintOptions(WithoutId()))
		assert.Equal(t, 1, len(*b.PrintOptions))
	})

	t.Run("WithPrintResults", func(t *testing.T) {
		b := Apply(WithPrintResults(WithoutId()))
		assert.Equal(t, 1, len(*b.PrintOptions))
	})

	t.Run("WithTransactionFolderName", func(t *testing.T) {
		b := Apply(WithTransactionFolderName("tx"))
		assert.Equal(t, "tx", b.TransactionFolderName)
	})

	t.Run("Overflow panics", func(t *testing.T) {
		assert.Panics(t, func() {
			Overflow(WithFlowConfig("nonexistant.json"))
		})
	})
}

func Apply(opt ...OverflowOption) *OverflowBuilder {
	return defaultOverflowBuilder.applyOptions(opt)
}

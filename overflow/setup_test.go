package overflow

import (
	"testing"

	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	t.Parallel()

	t.Run("default builder", func(t *testing.T) {
		o := NewOverflow()
		assert.Equal(t, output.InfoLog, o.LogLevel)
	})

	t.Run("default builder with loglevel from env", func(t *testing.T) {
		t.Setenv("OVERFLOW_LOGGING", "2")
		o := NewOverflow()
		assert.Equal(t, output.DebugLog, o.LogLevel)
	})

	t.Run("panic on wrong logging level", func(t *testing.T) {
		t.Setenv("OVERFLOW_LOGGING", "asd")
		assert.Panics(t, func() { NewOverflow() })
	})

	t.Run("none log", func(t *testing.T) {
		o := NewOverflow().NoneLog()
		assert.Equal(t, output.NoneLog, o.LogLevel)
	})

	t.Run("new overflow builder for network", func(t *testing.T) {
		o := NewOverflowForNetwork("testnet")
		assert.Equal(t, "testnet", o.Network)
	})

	t.Run("new overflow in memory", func(t *testing.T) {
		o := NewOverflowInMemoryEmulator()
		assert.Equal(t, "emulator", o.Network)
	})

	t.Run("new overflow testnet", func(t *testing.T) {
		o := NewOverflowTestnet()
		assert.Equal(t, "testnet", o.Network)
	})
	t.Run("new overflow mainnet", func(t *testing.T) {
		o := NewOverflowMainnet()
		assert.Equal(t, "mainnet", o.Network)
	})

	t.Run("new overflow emulator", func(t *testing.T) {
		o := NewOverflowEmulator()
		assert.Equal(t, "emulator", o.Network)
	})

	t.Run("new overflow builder without network", func(t *testing.T) {
		o := NewOverflowBuilder("", false, 1)
		assert.Equal(t, "emulator", o.Network)
	})

	t.Run("existing emulator", func(t *testing.T) {
		o := NewOverflow().ExistingEmulator()
		assert.Equal(t, false, o.DeployContracts)
		assert.Equal(t, false, o.InitializeAccounts)
	})

	t.Run("should panic on getting invalid account", func(t *testing.T) {
		o := NewOverflowInMemoryEmulator().ExistingEmulator().DoNotPrependNetworkToAccountNames()
		assert.Panics(t, func() { o.Start().Account("foobar") })
	})

	t.Run("do not prepend network names", func(t *testing.T) {
		o := NewOverflowInMemoryEmulator().ExistingEmulator().DoNotPrependNetworkToAccountNames()
		assert.Equal(t, false, o.PrependNetworkName)
		assert.Equal(t, "account", o.Start().ServiceAccountName())

	})

	t.Run("default gas", func(t *testing.T) {
		o := NewOverflow().DefaultGas(100)
		assert.Equal(t, 100, o.GasLimit)
	})

	t.Run("base path", func(t *testing.T) {
		o := NewOverflow().BasePath("../")
		assert.Equal(t, "../", o.Path)
	})

}

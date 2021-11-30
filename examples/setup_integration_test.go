package main

import (
	"github.com/bjartek/overflow/overflow"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetupIntegration(t *testing.T) {

	t.Run("Should create inmemory emulator client", func(t *testing.T) {
		g := overflow.NewGoWithTheFlowInMemoryEmulator()
		assert.Equal(t, "emulator", g.Network)
	})

	t.Run("Should create local emulator client", func(t *testing.T) {
		g := overflow.NewGoWithTheFlowEmulator()
		assert.Equal(t, "emulator", g.Network)
	})

	t.Run("Should create testnet client", func(t *testing.T) {
		g := overflow.NewGoWithTheFlowDevNet()
		assert.Equal(t, "testnet", g.Network)
	})

	t.Run("Should create testnet client with for network method", func(t *testing.T) {
		g := overflow.NewGoWithTheFlowForNetwork("testnet")
		assert.Equal(t, "testnet", g.Network)
	})

	t.Run("Should create mainnet client", func(t *testing.T) {
		g := overflow.NewGoWithTheFlowMainNet()
		assert.Equal(t, "mainnet", g.Network)
		assert.True(t, g.PrependNetworkToAccountNames)
		g = g.DoNotPrependNetworkToAccountNames()
		assert.False(t, g.PrependNetworkToAccountNames)

	})
}

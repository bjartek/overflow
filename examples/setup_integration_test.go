package main

import (
	"testing"

	"github.com/bjartek/overflow/overflow"
	"github.com/stretchr/testify/assert"
)

func TestSetupIntegration(t *testing.T) {

	t.Run("Should create inmemory emulator client", func(t *testing.T) {
		g := overflow.NewOverflowInMemoryEmulator().Start()
		assert.Equal(t, "emulator", g.Network)
	})
}

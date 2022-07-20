package overflow

import (
	"testing"

	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
)

func TestSetupIntegration(t *testing.T) {

	t.Run("Should create inmemory emulator client", func(t *testing.T) {
		g := NewOverflowInMemoryEmulator().Start()
		litter.Dump(g)
		assert.Equal(t, "emulator", g.Network)
	})
}

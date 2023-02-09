package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignIntegration(t *testing.T) {
	g, err := OverflowTesting()
	require.NoError(t, err)
	require.NotNil(t, g)

	t.Run("fail on missing signer", func(t *testing.T) {
		_, err := g.SignUserMessage("foobar", "baaaaaaaaanzaaaai")
		assert.ErrorContains(t, err, "could not find account with name emulator-foobar")
	})

	t.Run("should sign message", func(t *testing.T) {
		result, err := g.SignUserMessage("account", "overflow")
		assert.NoError(t, err)
		assert.Equal(t, 128, len(result))
	})

}

package overflow

import (
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
}

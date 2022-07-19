package overflow

import (
	"testing"

	"github.com/hexops/autogold"
	"github.com/sanity-io/litter"
	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	g := NewTestingEmulator().Start()
	t.Parallel()

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

	t.Run("parse and merge strip network prefix", func(t *testing.T) {
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
}

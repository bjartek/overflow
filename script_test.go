package overflow

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptArguments(t *testing.T) {
	g, err := NewTestingEmulator().StartE()
	require.NoError(t, err)
	t.Parallel()

	t.Run("Argument test", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.InlineScript("").Args(g.Arguments().UFix64(1.0))
		assert.Contains(t, builder.Arguments, ufix)
	})

	t.Run("Argument test values", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.InlineScript("").ArgsV(g.Arguments().UFix64(1.0).Build())
		assert.Contains(t, builder.Arguments, ufix)
	})

	t.Run("Argument test function", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.InlineScript("").ArgsFn(func(a *OverflowArgumentsBuilder) {
			a.UFix64(1.0)
		})
		assert.Contains(t, builder.Arguments, ufix)
	})
}

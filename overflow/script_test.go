package overflow

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
)

func TestScriptArguments(t *testing.T) {
	g := NewTestingEmulator().Start()
	t.Parallel()

	t.Run("Argument test", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.Script("").Args(g.Arguments().UFix64(1.0))
		assert.Contains(t, builder.Arguments, ufix)
	})

	t.Run("Argument test values", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.Script("").ArgsV(g.Arguments().UFix64(1.0).Build())
		assert.Contains(t, builder.Arguments, ufix)
	})

	t.Run("Argument test function", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.Script("").ArgsFn(func(a *FlowArgumentsBuilder) {
			a.UFix64(1.0)
		})
		assert.Contains(t, builder.Arguments, ufix)
	})
}

package overflow

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/stretchr/testify/assert"
)

/*
 Tests must be in the same folder as flow.json with contracts and transactions/scripts in subdirectories in order for the path resolver to work correctly
*/
func TestTransactionArguments(t *testing.T) {
	g := NewTestingEmulator().Start()
	t.Parallel()

	t.Run("Gas test", func(t *testing.T) {
		builder := g.Transaction("").Gas(100)
		assert.Equal(t, uint64(100), builder.GasLimit)
	})

	t.Run("Signer", func(t *testing.T) {
		builder := g.Transaction("").SignProposeAndPayAs("asd")
		assert.ErrorContains(t, builder.Error, "asd")
	})

	t.Run("Argument test builder", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.Transaction("").Args(g.Arguments().UFix64(1.0))
		assert.Contains(t, builder.Arguments, ufix)
	})

	t.Run("Argument test values", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.Transaction("").ArgsV(g.Arguments().UFix64(1.0).Build())
		assert.Contains(t, builder.Arguments, ufix)
	})

	t.Run("Argument test function", func(t *testing.T) {
		ufix, _ := cadence.NewUFix64("1.0")
		builder := g.Transaction("").ArgsFn(func(a *FlowArgumentsBuilder) {
			a.UFix64(1.0)
		})
		assert.Contains(t, builder.Arguments, ufix)
	})
}

package overflow

import (
	"testing"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-cli/pkg/flowkit/output"
	"github.com/stretchr/testify/assert"
)

func TestSetupFails(t *testing.T) {

	g := NewOverflow([]string{"../examples/flow.json"}, "emulator", true, output.NoneLog)
	_, err := g.CreateAccountsE("foobar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not find account with name foobar")

}

func TestScripArguments(t *testing.T) {

	g := NewOverflow([]string{"../examples/flow.json"}, "emulator", true, output.NoneLog)
	t.Run("Argument test builder", func(t *testing.T) {
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

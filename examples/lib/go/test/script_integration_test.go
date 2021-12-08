package main

import (
	"testing"

	"github.com/bjartek/overflow/overflow"
	"github.com/stretchr/testify/assert"
)

func TestScriptSubdir(t *testing.T) {
	g := overflow.NewTestingEmulator()
	t.Parallel()

	t.Run("Raw account argument", func(t *testing.T) {
		value := g.ScriptFromFile("test").ScriptPath("../../../scripts/").RawAccountArgument("0x1cf0e2f2f715450").RunReturnsInterface()
		assert.Equal(t, "0x1cf0e2f2f715450", value)
	})

}

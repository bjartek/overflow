package main

import (
	"testing"

	"github.com/bjartek/overflow/overflow"
	"github.com/stretchr/testify/assert"
)

func TestScriptSubdir(t *testing.T) {
	g := overflow.NewTestingEmulator().Start()
	t.Parallel()

	t.Run("Raw account argument", func(t *testing.T) {
		value := g.ScriptFromFile("test").
			ScriptPath("../../../scripts/").
			Args(g.Arguments().RawAccount("0x1cf0e2f2f715450")).
			RunReturnsInterface()
		assert.Equal(t, "0x1cf0e2f2f715450", value)
	})

}

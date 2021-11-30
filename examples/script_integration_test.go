package main

import (
	"github.com/bjartek/go-with-the-flow/v2/gwtf"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestScript(t *testing.T) {
	g := gwtf.NewTestingEmulator()
	t.Parallel()

	t.Run("Raw account argument", func(t *testing.T) {
		value := g.ScriptFromFile("test").RawAccountArgument("0x1cf0e2f2f715450").RunReturnsInterface()
		assert.Equal(t, "0x1cf0e2f2f715450", value)
	})

	t.Run("Raw account argument", func(t *testing.T) {
		value := g.ScriptFromFile("test").AccountArgument("first").RunReturnsInterface()
		assert.Equal(t, "0x1cf0e2f2f715450", value)
	})

	t.Run("Script should report failure", func(t *testing.T) {
		value, err := g.Script("asdf").RunReturns()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Parsing failed")
		assert.Nil(t, value)

	})

}

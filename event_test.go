package overflow

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvent(t *testing.T) {

	g, err := NewTestingEmulator().StartE()
	require.NoError(t, err)

	t.Run("Start argument", func(t *testing.T) {
		ef := g.EventFetcher().Start(100)
		assert.Equal(t, ef.FromIndex, int64(100))
	})

	t.Run("From argument", func(t *testing.T) {
		ef := g.EventFetcher().From(100)
		assert.Equal(t, ef.FromIndex, int64(100))
	})

	t.Run("End argument", func(t *testing.T) {
		ef := g.EventFetcher().End(100)
		assert.Equal(t, ef.EndIndex, uint64(100))
		assert.False(t, ef.EndAtCurrentHeight)
	})

	t.Run("Until argument", func(t *testing.T) {
		ef := g.EventFetcher().Until(100)
		assert.Equal(t, ef.EndIndex, uint64(100))
		assert.False(t, ef.EndAtCurrentHeight)
	})

	t.Run("Until current argument", func(t *testing.T) {
		ef := g.EventFetcher().UntilCurrent()
		assert.Equal(t, ef.EndIndex, uint64(0))
		assert.True(t, ef.EndAtCurrentHeight)
	})

	t.Run("workers argument", func(t *testing.T) {
		ef := g.EventFetcher().Workers(100)
		assert.Equal(t, ef.NumberOfWorkers, 100)
	})

	t.Run("batch size argument", func(t *testing.T) {
		ef := g.EventFetcher().BatchSize(100)
		assert.Equal(t, ef.EventBatchSize, uint64(100))
	})

	t.Run("event ignoring field argument", func(t *testing.T) {
		ef := g.EventFetcher().EventIgnoringFields("foo", []string{"bar", "baz"})
		assert.Equal(t, ef.EventsAndIgnoreFields["foo"], []string{"bar", "baz"})
	})

	t.Run("failed reading invalid file", func(t *testing.T) {
		_, err := readProgressFromFile("boing.boinb")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ProgressFile is not valid open boing.boinb")
	})

	t.Run("Cannot write to file that is dir", func(t *testing.T) {
		err := os.Mkdir("foo", os.ModeDir)
		assert.NoError(t, err)
		defer os.RemoveAll("foo")

		err = writeProgressToFile("foo", 1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "foo: is a directory")

	})
}

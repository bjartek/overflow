package overflow

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationEvents(t *testing.T) {

	t.Run("Test that from index cannot be negative", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NoError(t, err)

		g.Tx("mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.0),
		).AssertSuccess(t).
			AssertEventCount(t, 3)

		_, err = g.FetchEvents(
			WithEndIndex(2),
			WithFromIndex(-10),
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FromIndex is negative")
	})

	t.Run("Fetch last events", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NoError(t, err)
		g.Tx("mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.0),
		).AssertSuccess(t).
			AssertEventCount(t, 3)
		ev, err := g.FetchEvents(WithLastBlocks(2), WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
	})

	t.Run("Fetch last events and sort them ", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NoError(t, err)
		g.Tx("mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.0),
		).AssertSuccess(t).
			AssertEventCount(t, 3)

		g.Tx("mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.0),
		).AssertSuccess(t).
			AssertEventCount(t, 3)

		ev, err := g.FetchEvents(WithLastBlocks(2), WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"))
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ev))
		assert.True(t, ev[0].BlockHeight < ev[1].BlockHeight)
	})

	t.Run("Fetch last write progress file", func(t *testing.T) {
		g, err := OverflowTesting()
		require.NoError(t, err)
		g.Tx("mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.0),
		).AssertSuccess(t).
			AssertEventCount(t, 3)

		ev, err := g.FetchEvents(WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"), WithTrackProgressIn("progress"))
		defer os.Remove("progress")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
	})

	t.Run("should fail reading invalid progress from file", func(t *testing.T) {
		err := os.WriteFile("progress", []byte("invalid"), fs.ModePerm)
		assert.NoError(t, err)

		g, err := OverflowTesting()
		require.NoError(t, err)

		_, err = g.FetchEvents(WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"), WithTrackProgressIn("progress"))
		defer os.Remove("progress")
		assert.Error(t, err)
		assert.Equal(t, "could not parse progress file as block height strconv.ParseInt: parsing \"invalid\": invalid syntax", err.Error())
	})

	t.Run("Fetch last write progress file that exists", func(t *testing.T) {

		err := os.WriteFile("progress", []byte("1"), fs.ModePerm)
		assert.NoError(t, err)

		g, err := OverflowTesting()
		assert.NoError(t, err)
		g.Tx("mint_tokens",
			WithSignerServiceAccount(),
			WithArg("recipient", "first"),
			WithArg("amount", 100.0),
		).AssertSuccess(t).
			AssertEventCount(t, 3)

		ev, err := g.FetchEvents(WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"), WithTrackProgressIn("progress"))
		defer os.Remove("progress")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(ev))
	})
}

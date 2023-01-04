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
		g, err := NewTestingEmulator().StartE()
		require.NoError(t, err)
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			Args(g.Arguments().
				Account("first").
				UFix64(100.0)).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		_, err = g.EventFetcher().End(2).From(-10).Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").Run()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FromIndex is negative")
	})

	t.Run("Fetch last events", func(t *testing.T) {

		g, err := NewTestingEmulator().StartE()
		require.NoError(t, err)
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			Args(g.Arguments().
				Account("first").
				UFix64(100.0)).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		ev, err := g.EventFetcher().Last(2).Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").Run()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
	})

	t.Run("Fetch last events and sort them ", func(t *testing.T) {

		g, err := NewTestingEmulator().StartE()
		require.NoError(t, err)
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			Args(g.Arguments().
				Account("first").
				UFix64(100.0)).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			Args(g.Arguments().
				Account("first").
				UFix64(100.0)).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		ev, err := g.EventFetcher().Last(3).Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").Run()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ev))
		assert.True(t, ev[0].BlockHeight < ev[1].BlockHeight)
	})

	t.Run("Fetch last write progress file", func(t *testing.T) {
		g, err := NewTestingEmulator().StartE()
		require.NoError(t, err)
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			Args(g.Arguments().
				Account("first").
				UFix64(100.0)).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		ev, err := g.EventFetcher().Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").TrackProgressIn("progress").Run()
		defer os.Remove("progress")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
	})

	t.Run("should fail reading invalid progress from file", func(t *testing.T) {
		err := os.WriteFile("progress", []byte("invalid"), fs.ModePerm)
		assert.NoError(t, err)

		g, err := NewTestingEmulator().StartE()
		require.NoError(t, err)
		_, err = g.EventFetcher().Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").TrackProgressIn("progress").Run()
		defer os.Remove("progress")
		assert.Error(t, err)
		assert.Equal(t, "could not parse progress file as block height strconv.ParseInt: parsing \"invalid\": invalid syntax", err.Error())
	})

	t.Run("Fetch last write progress file that exists", func(t *testing.T) {

		err := os.WriteFile("progress", []byte("1"), fs.ModePerm)
		assert.NoError(t, err)

		g, err := NewTestingEmulator().StartE()
		require.NoError(t, err)
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			Args(g.Arguments().
				Account("first").
				UFix64(100.0)).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		ev, err := g.EventFetcher().Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").TrackProgressIn("progress").Run()
		defer os.Remove("progress")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
	})
}

package main

import (
	"io/fs"
	"os"
	"testing"

	"github.com/bjartek/overflow/overflow"
	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {

	t.Run("Test that from index cannot be negative", func(t *testing.T) {
		g := overflow.NewTestingEmulator().Start()
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			AccountArgument("first").
			UFix64Argument(100.0).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		_, err := g.EventFetcher().End(2).From(-10).Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").Run()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FromIndex is negative")
	})

	t.Run("Fetch last events", func(t *testing.T) {
		g := overflow.NewTestingEmulator().Start()
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			AccountArgument("first").
			UFix64Argument(100.0).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		ev, err := g.EventFetcher().Last(2).Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").Run()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
	})

	t.Run("Fetch last events and sort them ", func(t *testing.T) {
		g := overflow.NewTestingEmulator().Start()
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			AccountArgument("first").
			UFix64Argument(100.0).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			AccountArgument("first").
			UFix64Argument(100.0).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		ev, err := g.EventFetcher().Last(3).Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").Run()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ev))
		assert.True(t, ev[0].BlockHeight < ev[1].BlockHeight)
	})

	t.Run("Fetch last write progress file", func(t *testing.T) {
		g := overflow.NewTestingEmulator().Start()
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			AccountArgument("first").
			UFix64Argument(100.0).
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

		g := overflow.NewTestingEmulator().Start()
		_, err = g.EventFetcher().Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").TrackProgressIn("progress").Run()
		defer os.Remove("progress")
		assert.Error(t, err)
		assert.Equal(t, "could not parse progress file as block height strconv.ParseInt: parsing \"invalid\": invalid syntax", err.Error())
	})

	t.Run("Fetch last write progress file that exists", func(t *testing.T) {

		err := os.WriteFile("progress", []byte("1"), fs.ModePerm)
		assert.NoError(t, err)

		g := overflow.NewTestingEmulator().Start()
		g.TransactionFromFile("mint_tokens").
			SignProposeAndPayAsService().
			AccountArgument("first").
			UFix64Argument(100.0).
			Test(t).
			AssertSuccess().
			AssertEventCount(3)

		ev, err := g.EventFetcher().Event("A.0ae53cb6e3f42a79.FlowToken.TokensMinted").TrackProgressIn("progress").Run()
		defer os.Remove("progress")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
	})

}

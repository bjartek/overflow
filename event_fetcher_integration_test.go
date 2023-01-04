package overflow

import (
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startOverflowAndMintTokens(t *testing.T) *OverflowState {
	t.Helper()
	o, err := OverflowTesting()
	require.NoError(t, err)
	result := o.Tx("mint_tokens", WithSignerServiceAccount(), WithArg("recipient", "first"), WithArg("amount", 100.0))
	assert.NoError(t, result.Err)
	return o

}

type MarketEvent struct {
	EventDate         time.Time `json:"eventDate"`
	FlowEventID       string    `json:"flowEventId"`
	FlowTransactionID string    `json:"flowTransactionId"`
	ID                string    `json:"id"`
	BlockEventData    struct {
		Amount float64 `json:"amount"`
	} `json:"blockEventData"`
}

func TestIntegrationEventFetcher(t *testing.T) {

	t.Run("Test that from index cannot be negative", func(t *testing.T) {
		_, err := startOverflowAndMintTokens(t).FetchEvents(
			WithEndIndex(2),
			WithFromIndex(-10),
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FromIndex is negative")
	})

	t.Run("Fetch last events", func(t *testing.T) {
		ev, err := startOverflowAndMintTokens(t).FetchEvents(
			WithLastBlocks(2),
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
		)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
	})

	t.Run("Fetch last events and sort them ", func(t *testing.T) {
		o := startOverflowAndMintTokens(t)
		result := o.Tx("mint_tokens", WithSignerServiceAccount(), WithArg("recipient", "first"), WithArg("amount", "100.0"))
		assert.NoError(t, result.Err)
		ev, err := o.FetchEvents(
			WithLastBlocks(3),
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
		)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(ev))
		assert.True(t, ev[0].BlockHeight < ev[1].BlockHeight)
	})

	t.Run("Fetch last write progress file", func(t *testing.T) {
		ev, err := startOverflowAndMintTokens(t).FetchEvents(
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
			WithTrackProgressIn("progress"),
		)
		defer os.Remove("progress")
		assert.NoError(t, err)
		assert.Equal(t, 1, len(ev))
		assert.Contains(t, ev[0].String(), "100")
	})

	t.Run("should fail reading invalid progress from file", func(t *testing.T) {
		err := os.WriteFile("progress", []byte("invalid"), fs.ModePerm)
		assert.NoError(t, err)

		_, err = startOverflowAndMintTokens(t).FetchEvents(
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
			WithTrackProgressIn("progress"),
		)
		defer os.Remove("progress")
		assert.Error(t, err)
		assert.Equal(t, "could not parse progress file as block height strconv.ParseInt: parsing \"invalid\": invalid syntax", err.Error())
	})

	t.Run("Fetch last write progress file that exists and marshal events", func(t *testing.T) {

		err := os.WriteFile("progress", []byte("1"), fs.ModePerm)
		assert.NoError(t, err)

		ev, err := startOverflowAndMintTokens(t).FetchEvents(
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
			WithTrackProgressIn("progress"),
		)
		defer os.Remove("progress")
		assert.NoError(t, err)
		assert.Equal(t, 3, len(ev))
		event := ev[0]

		graffleEvent := event.ToGraffleEvent()

		var eventMarshal map[string]interface{}
		assert.NoError(t, event.MarshalAs(&eventMarshal))
		assert.NotEmpty(t, eventMarshal)

		autogold.Equal(t, graffleEvent.BlockEventData, autogold.Name("graffle-event"))
		var marshalTo MarketEvent
		assert.NoError(t, graffleEvent.MarshalAs(&marshalTo))
		assert.Equal(t, float64(10), marshalTo.BlockEventData.Amount)
	})

	t.Run("Return progress writer ", func(t *testing.T) {
		progressFile := "progress"

		err := writeProgressToFile(progressFile, 0)
		require.NoError(t, err)
		res := startOverflowAndMintTokens(t).FetchEventsWithResult(
			WithEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted"),
			WithTrackProgressIn(progressFile),
			WithReturnProgressWriter(),
		)
		require.NoError(t, res.Error)
		progress, err := readProgressFromFile(progressFile)
		require.NoError(t, err)
		assert.Equal(t, int64(0), progress)

		res.ProgressWriteFunction()

		progress, err = readProgressFromFile(progressFile)
		require.NoError(t, err)
		assert.Equal(t, int64(9), progress)

		ev := res.Events
		defer os.Remove(progressFile)
		assert.Equal(t, 1, len(ev))
		assert.Contains(t, ev[0].String(), "100")
	})

}

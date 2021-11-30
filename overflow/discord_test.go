package overflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscordHooks(t *testing.T) {

	t.Run("Should create hook", func(t *testing.T) {

		hook := NewDiscordWebhook("http://foo/bar/123/456")
		assert.Equal(t, "123", hook.ID)
		assert.Equal(t, "456", hook.Token)

	})

	t.Run("Should parse message", func(t *testing.T) {

		ev := NewTestEvent("A.0ae53cb6e3f42a79.FlowToken.TokensMinted", map[string]interface{}{"amount": "100.00000000"})

		message := EventsToWebhookParams([]*FormatedEvent{ev})
		embedded := message.Embeds[0]
		assert.Equal(t, "A.0ae53cb6e3f42a79.FlowToken.TokensMinted", embedded.Title)
		assert.Equal(t, "amount", embedded.Fields[0].Name)
		assert.Equal(t, "100.00000000", embedded.Fields[0].Value)

	})
}

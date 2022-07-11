package overflow

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//TODO: look at updating the discord go dependency

// DiscordWebhook stores information about a webhook
type DiscordWebhook struct {
	ID    string `json:"id"`
	Token string `json:"token"`
	Wait  bool   `json:"wait"`
}

//NewDiscordWebhook create a new discord webhook from an discord url on the form ofhttps://discord.com/api/webhooks/<id>/<token>
func NewDiscordWebhook(url string) DiscordWebhook {
	parts := strings.Split(url, "/")
	length := len(parts)
	return DiscordWebhook{
		ID:    parts[length-2],
		Token: parts[length-1],
		Wait:  true,
	}
}

// SendEventsToWebhook Sends events to a webhook
func (dw DiscordWebhook) SendEventsToWebhook(events []*FormatedEvent) (*discordgo.Message, error) {
	discord, err := discordgo.New()
	if err != nil {
		return nil, err
	}

	status, err := discord.WebhookExecute(
		dw.ID,
		dw.Token,
		dw.Wait,
		EventsToWebhookParams(events))

	if err != nil {
		return nil, err
	}
	return status, nil
}

//EventsToWebhookParams convert events to rich webhook
func EventsToWebhookParams(events []*FormatedEvent) *discordgo.WebhookParams {
	var embeds []*discordgo.MessageEmbed
	for _, event := range events {

		var fields []*discordgo.MessageEmbedField
		for name, value := range event.Fields {
			fields = append(fields, &discordgo.MessageEmbedField{
				Name:  name,
				Value: fmt.Sprintf("%v", value),
			})
		}

		embeds = append(embeds, &discordgo.MessageEmbed{
			Title:  event.Name,
			Type:   discordgo.EmbedTypeRich,
			Fields: fields,
			Footer: &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("blockHeight %d @ %s", event.BlockHeight, event.Time),
			},
		})
	}

	return &discordgo.WebhookParams{
		Embeds: embeds,
	}
}

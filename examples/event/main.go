package main

import (
	"log"

	"github.com/bjartek/overflow/overflow"
)

func main() {

	g := overflow.NewOverflowTestnet().Start()

	eventsFetcher := g.EventFetcher().
		Last(1000).
		Event("A.0b2a3299cc857e29.TopShot.Withdraw")
		//EventIgnoringFields("A.0b2a3299cc857e29.TopShot.Withdraw", []string{"field1", "field"})

	events, err := eventsFetcher.Run()
	if err != nil {
		panic(err)
	}

	log.Printf("%v", events)

	//to send events to a discord eventhook use
	//	message, err := overflow.NewDiscordWebhook("http://your-webhook-url").SendEventsToWebhook(events)

}

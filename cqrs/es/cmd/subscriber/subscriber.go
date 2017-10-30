package main

import (
	"encoding/json"

	"os"
	"os/signal"

	"github.com/sokool/gokit/cqrs/es/client"
	"github.com/sokool/gokit/log"
)

type CreatedEvent struct {
	Name    string
	Address string
}

type SubscribedEvent struct {
	Person string
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	//streamName := "restaurant"
	//streamID := "1312"

	subscriber, err := client.NewWebSocket().NewSubscriber()

	if err != nil {
		log.Fatal("cmd.client.reader", err)
	}

	go func() {
		<-signals
		subscriber.Close()
	}()

	subscriber.Read(func(ms []client.Event) {
		for _, m := range ms {
			var e interface{}
			switch m.EventName {
			case "CreatedEvent":
				e = &CreatedEvent{}
			case "SubscribedEvent":
				e = &SubscribedEvent{}
			}

			json.Unmarshal(m.Data, e)

			log.Info("cmd.client.reader", "v.%d.%s.%s%+v, Meta:%+v",
				m.Version, m.TopicName, m.EventName, e, string(m.Meta))
		}
	})

}

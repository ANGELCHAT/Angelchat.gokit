package main

import (
	"time"

	"os"
	"os/signal"

	"fmt"

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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	streamName := "restaurant"
	streamID := "1312"
	metaData := map[string]string{
		"ip":      "125.99.21.31",
		"user":    "Mike",
		"session": "941ASDkas1412314dAdF"}

	writer, err := client.NewWebSocket().NewWriter(streamName)
	if err != nil {
		log.Fatal("cmd.client.writer", err)
	}

	go func() {
		s := <-sigs
		log.Debug("cmd.client.writer", "signal %s received", s.String())
		writer.Close()
	}()

	events := []interface{}{
		CreatedEvent{"PasiBus", "Legnicka 52, Wrocław"},
		CreatedEvent{"Mamma Mia", "Piłsudskiego 13, Wrocław"},
		CreatedEvent{"Mango Mama", "Jedności Narodowej 31, Wrocław"},
		SubscribedEvent{"Tomek"},
		SubscribedEvent{"Zygmunt"},
		SubscribedEvent{"Albert"},
		SubscribedEvent{"Katarzyna"},
		SubscribedEvent{"Michalina"},
		SubscribedEvent{"Anna"},
		SubscribedEvent{"Grzegorz"},
		SubscribedEvent{"Franek"},
	}

	for range time.NewTicker(1500 * time.Millisecond).C {
		log.Debug("cmd.client.writer", "sending...")
		if err := writer.Write(streamID, metaData, events...); err != nil {
			log.Fatal("cmd.client.writer", fmt.Errorf("sending events to %s", err.Error()))
		}
	}

	time.Sleep(time.Second * 3)
	log.Debug("cmd.client.writer", "exit")

}

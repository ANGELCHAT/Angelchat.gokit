package socket

import (
	"context"
	"fmt"
)

type message struct {
	Type          string `json:"@type"`
	CorrelationID string `json:"@correlationId"`
}

type Message struct {
	message
	Body []byte //io.ReadCloser

	Termination context.Context
	Closed      bool
	stream      chan<- []byte
}

func (m *Message) Decode(v interface{}) error {
	return decode(m.Body, v)
}

func (m *Message) Respond(typ string, data interface{}) error {
	return m.Create(m.CorrelationID, typ, data)
}

func (m *Message) Create(id, typ string, data interface{}) error {
	b, err := encode(message{typ, id}, data)
	if err != nil {
		return err
	}

	select {
	case <-m.Termination.Done():
		return fmt.Errorf("disconnected by peer")
	case m.stream <- b:

	}

	return nil
}

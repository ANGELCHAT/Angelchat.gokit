package socket

import (
	"bufio"
	"context"
	"io"

	"github.com/google/uuid"
)

func NewConnection(p *Peer, ctx context.Context) *Connection {
	messages := make(chan Message)

	go func() { // read socket and transform raw bytes in Message
		defer close(messages)

		var (
			socket  = bufio.NewReader(p)
			err     error
			raw     []byte
			decoded message
		)

		for { // read until reader receive end of file
			if raw, err = socket.ReadBytes('\n'); err != nil && err == io.EOF {
				return
			}

			if err := decode(raw, &decoded); err != nil {
				continue
			}

			messages <- Message{decoded, raw}
		}
	}()

	return &Connection{
		Termination: ctx,
		Messages:    messages,
		socket:      p,
	}
}
// Transmission
type Connection struct {
	Termination context.Context
	Messages    <-chan Message
	socket      io.ReadWriter
}

func (s *Connection) Write(typ, cid string, m interface{}) error {
	raw, err := encode(message{uuid.New().String(), cid, typ}, m)
	if err != nil {
		return err
	}

	_, err = s.socket.Write(raw)

	return err
}

type Message struct {
	message
	Body []byte `json:"-"`
}

func (m Message) Decode(v interface{}) error { return decode(m.Body, v) }

type message struct {
	ID   string `json:"@id"`   // unique message id, each message has different
	CID  string `json:"@cid"`  // correlation ID between processes
	Type string `json:"@type"` // just a type of message
}

type Handler interface {
	ServeWS(*Connection)
}

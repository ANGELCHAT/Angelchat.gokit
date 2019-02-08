package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/teris-io/shortid"
)

type Handler interface {
	ServeWS(*Message) error
}

type Server struct {
	log         logger
	counter     *uint64
	shutdown    context.Context
	receiver    Handler
	websocket   websocket.Upgrader
	connections sync.Map
}

func NewServer(h Handler, options ...option) *Server {
	o, _ := newOptions(options...)

	var messages uint64 = 0
	return &Server{
		log:      o.log,
		counter:  &messages,
		receiver: h,
		shutdown: o.shutdown,
		websocket: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (c *Server) Connect(w http.ResponseWriter, r *http.Request) {
	var (
		id                    = shortid.MustGenerate()
		err                   error
		log                   = c.log.WithTag("Websocket." + id)
		client                *websocket.Conn
		responses             = make(chan []byte)
		sentinel              sync.WaitGroup
		disconnection, cancel = context.WithCancel(context.Background())
	)

	defer close(responses)

	if client, err = c.websocket.Upgrade(w, r, nil); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Info("%s peer connected", r.RemoteAddr)

	// receive close frames from client socket
	client.SetCloseHandler(func(code int, text string) error {
		log.Debug("close frame received from socket %d:%s", code, text)

		cancel()
		return nil
	})

	//create and configure Request structure
	command := func(b []byte) *Message {
		var m message
		if err := decode(b, &m); err != nil {
			log.Info("command: decode failed due %s", err)
			//todo send message to client
			return nil
		}

		return &Message{
			message:     m,
			Body:        b,
			Termination: disconnection,
			stream:      responses,
		}
	}

	sentinel.Add(1)
	go func() {
		defer cancel()
		defer client.Close()
		defer log.Debug("closer: goroutine finished")
		defer sentinel.Done()

		select {
		case <-disconnection.Done():
			log.Debug("closer: disconnected signal received")

		case <-c.shutdown.Done():
			log.Debug("closer: shutdown signal received")
		}

		t := time.Now().Add(time.Second)
		m := websocket.FormatCloseMessage(websocket.CloseGoingAway, "flux.server shutting down")
		if err := client.WriteControl(websocket.CloseMessage, m, t); err != nil {
			log.Debug("closer: sending close message to peer failed due %s", err)
			return
		}

		log.Debug("closer: send close message to peer")
	}()

	// write take care of deliver messages to websocket client, it finished
	// when websocket client not responding or when done channel is closed.
	write := func(socket *websocket.Conn) {
		sentinel.Add(1)
		ping := time.NewTicker((time.Second * 10 * 9) / 10)

		go func() {
			defer ping.Stop()
			defer cancel()
			defer log.Debug("write: goroutine finished")
			defer sentinel.Done()

			for {
				select {
				case <-disconnection.Done():
					return

				case <-ping.C:
					// ping client to check if connection is still established.
					if err := socket.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Debug("write: ping failed due %s", err.Error())
						return
					}

				case r, ok := <-responses:
					if !ok { // responses has been closed - finish write goroutine
						return
					}

					//log.Debug("write: sending %+v", r)
					if err := socket.WriteMessage(websocket.TextMessage, r); err != nil {
						// TODO
						//  list of error types, decide if loop should be finish
						//  here. That's depend on types of errors from write
						//  read. List below:
						//    - net.OpError
						//	  - json.
						log.Debug("write: failed due %s [%T]", err.Error(), err)
						return
					}

					// todo
					//  Implement responses logging with ability to stack
					//  multiple responses and combine it in one debug message.
					//log.Debug("<- %s[%s]", res.Name, res.Channel)

					//
					atomic.AddUint64(c.counter, 1)
				}
			}
		}()
	}

	// read messages from websocket and writes them into messages channel,
	// method finish when client is disconnected or when done channel is closed.
	read := func(socket *websocket.Conn) {
		sentinel.Add(1)
		go func() {
			defer cancel()
			defer log.Debug("read: goroutine finished")
			defer sentinel.Done()

			for { // until error from websocket
				_, message, err := socket.ReadMessage()
				if err != nil {
					log.Debug("read: socket failed due %s [%T]", err.Error(), err)
					return
				}

				//r := command(message)

				sentinel.Add(1)
				go func(b []byte) {
					m := command(b)

					defer log.Debug("handler: [%s] goroutine finished", m.Type)
					defer sentinel.Done()

					c.receiver.ServeWS(m)
				}(message)

				atomic.AddUint64(c.counter, 1)
			}
		}()
	}

	read(client)
	write(client)

	responses <- []byte(fmt.Sprintf(`{"@type":"Connected", "@correlationId":"%s"}`, id))

	c.connections.Store(id, cancel)

	sentinel.Wait()

	log.Info("%s peer disconnected", r.RemoteAddr)
}

func decode(what []byte, where ...interface{}) error {
	for i := range where {
		if err := json.Unmarshal(what, where[i]); err != nil {
			return err
		}
	}

	return nil
}

func encode(what ...interface{}) ([]byte, error) {
	var d map[string]interface{}

	for i := range what {
		mb, err := json.Marshal(what[i])
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(mb, &d); err != nil {
			return nil, err
		}
	}

	return json.Marshal(d)
}

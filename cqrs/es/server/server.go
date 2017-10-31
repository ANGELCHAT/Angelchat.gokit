package server

import (
	"net/http"

	"fmt"

	"context"
	"time"

	"sync"

	"github.com/gorilla/websocket"
	"github.com/sokool/gokit/cqrs/es"
	"github.com/sokool/gokit/log"
)

var tag = fmt.Sprintf("es.server")

type Server struct {
	http *http.Server
	opts *options
	//
	done chan struct{}
	// hold all registered actions
	actions sync.Map
}

func (s *Server) serve(f func(*es.Service) TransmitReceiver) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {

		ws, err := websocket.Upgrade(res, req, nil, 1024, 1024)
		if err != nil {
			log.Error(tag, fmt.Errorf("websocket handshake %s", err))
			return
		}

		log.Info(tag, "%s connected", req.RemoteAddr)

		action := newAction(ws)

		s.actions.Store(action, true)
		action.do(f(s.opts.eventStore))
		s.actions.Delete(action)

		log.Info(tag, "%s disconnected", req.RemoteAddr)

		var length int
		s.actions.Range(func(_, _ interface{}) bool {
			length++
			return true
		})

		if length == 0 {
			close(s.done)
		}
	}
}

func (s *Server) Start() error {

	m := http.NewServeMux()
	m.HandleFunc("/stream", s.serve(newStreamer))
	m.HandleFunc("/subscribe", s.serve(newSubscriber))

	s.http = &http.Server{Addr: s.opts.address, Handler: m}

	log.Info(tag, "listening on http://%s", s.opts.address)
	if err := s.http.ListenAndServe(); err != nil {
		//log.Info(tag, "server closed %T %s", err, err.Error())
	}

	<-s.done

	return nil
}

// Close send close message to all clients in order to let them know that
// server is going to be shutdown.
func (s *Server) Close() error {
	s.actions.Range(func(key, value interface{}) bool {
		a, ok := key.(*action)
		if !ok {
			log.Error(tag, fmt.Errorf("wrong key in actions"))
			return true
		}
		a.shutdown <- "server shutdown"

		return true
	})

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	return s.http.Shutdown(ctx)

}

func New(opts ...Option) *Server {
	return &Server{
		opts: newOptions(opts...),
		done: make(chan struct{}),
	}

}

type Data struct {
	Bytes []byte
	Error error
}

type TransmitReceiver interface {
	Transmit() <-chan Data
	Receive([]byte) error
	Close() error
}

type action struct {
	conn     *websocket.Conn
	shutdown chan string
}

func (a *action) do(t TransmitReceiver) {
	done := make(chan string)
	write := make(chan []byte)
	go func() {
		for m := range t.Transmit() {
			if m.Error != nil {
				log.Error(tag, fmt.Errorf("reading from handler: %s", m.Error.Error()))
				continue
			}

			write <- m.Bytes
		}
		close(write)
		done <- "transmission interrupted"

	}()
	read := make(chan []byte)
	go func() {
		defer close(read)
		for {
			_, b, err := a.conn.ReadMessage()
			if err != nil { // when close message received, unregister action
				a.handleError(err, "read")
				return
			}
			read <- b
		}

	}()

	defer func() {
		if err := t.Close(); err != nil {
			log.Error(tag, err)
		}

		if err := a.conn.Close(); err != nil {
			log.Error(tag, err)
		}
	}()

	for {
		select {
		case s, ok := <-done:
			if !ok {
				log.Debug(tag, "done:waiting for read closing")
				time.Sleep(time.Millisecond * 100)
				continue
			}

			a.sendCloseMessage(s)

		case s, ok := <-a.shutdown:
			if !ok {
				log.Debug(tag, "shutdown:waiting for read closing")
				time.Sleep(time.Millisecond * 100)
				continue
			}
			a.sendCloseMessage(s)
			close(a.shutdown)

		case b, ok := <-write:
			if !ok {
				log.Debug(tag, "write:waiting for read closing")
				time.Sleep(time.Millisecond * 100)
				continue
			}

			err := a.conn.WriteMessage(websocket.TextMessage, b)
			if err != nil {
				a.handleError(err, "write")
				return
			}

		case b, ok := <-read: // channel is closed, when websocket close message arrive.
			if !ok {
				log.Debug(tag, "%s receiving done", a.conn.RemoteAddr())
				return // method action.do is finished
			}

			if err := t.Receive(b); err != nil {
				log.Error(tag, fmt.Errorf("writing to handler %s", err.Error()))
			}
		}
	}
}
func (a *action) Close() error {

	return nil
}

func (a *action) sendCloseMessage(s string) {
	msg := websocket.FormatCloseMessage(websocket.CloseGoingAway, s)
	if err := a.conn.WriteMessage(websocket.CloseMessage, msg); err != nil {
		log.Error(tag, err)
	}
	log.Debug(tag, "%s '%s' close message send to peer", a.conn.RemoteAddr(), s)
}

func (a *action) handleError(err error, s string) {
	if err == nil {
		return
	}

	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) { // it's ok
		log.Debug(tag, "%s %s, received close message from peer [%s]",
			a.conn.RemoteAddr().String(), s, err.Error())
	} else if err != nil {
		log.Error(tag, fmt.Errorf("%s peer close received [%s]", s, err.Error()))
	}

}

func newAction(c *websocket.Conn) *action {
	return &action{
		conn:     c,
		shutdown: make(chan string),
	}
}

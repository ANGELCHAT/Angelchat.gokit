package server

import (
	"context"
	"net/http"

	"time"

	"fmt"

	"sync"

	"github.com/gorilla/websocket"
	"github.com/sokool/gokit/cqrs/es"
	"github.com/sokool/gokit/log"
)

type Server struct {
	sync.Mutex
	http *http.Server
	opts *options
	//todo thread safe map?
	connections map[string]*websocket.Conn
	done        chan struct{}
}

func (s *Server) serve(f func(*es.Service) TransmitReceiver) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		tr := f(s.opts.eventStore)
		tag := fmt.Sprintf("es.server")
		wsc, err := websocket.Upgrade(res, req, nil, 1024, 1024)
		if err != nil {
			log.Error(tag, fmt.Errorf("websocket handshake %s", err))
			return
		}
		//done := make(chan struct{})

		s.Lock()
		s.connections[req.RemoteAddr] = wsc
		s.Unlock()

		go func() { // read from handler(blocks) and write to peer
			defer func() {
				log.Debug(tag, "writing to peer done")
			}()
			for d := range tr.Transmit() {
				log.Info(tag, "writing to peer")
				if d.Error != nil {
					log.Error(tag, fmt.Errorf("reading from handler: %s", d.Error.Error()))
					continue
				}

				err = wsc.WriteMessage(websocket.TextMessage, d.Bytes)
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) { // it's ok
					return
				} else if err != nil {
					log.Error(tag, fmt.Errorf("writing to peer %s", err.Error()))
					return
				}
			}
		}()

		for {
			_, b, err := wsc.ReadMessage()
			log.Info(tag, "read from peer", "%s", string(b))
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) { // it's ok
				break
			} else if err != nil { // unexpected error
				log.Error(tag, fmt.Errorf("reading from peer %s", err.Error()))
				break
			}

			if err := tr.Receive(b); err != nil {
				log.Error(tag, fmt.Errorf("writing to handler %s", err.Error()))
			}
		}

		wsc.Close()
		log.Debug(tag, "connection closed")
	}
}

func (s *Server) serve2(f func(*es.Service) TransmitReceiver) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		tr := f(s.opts.eventStore)
		done := make(chan struct{})
		tag := fmt.Sprintf("es.server")
		wsc, err := websocket.Upgrade(res, req, nil, 1024, 1024)
		if err != nil {
			log.Error(tag, fmt.Errorf("websocket handshake %s", err))
			return
		}
		s.Lock()
		s.connections[req.RemoteAddr] = wsc
		s.Unlock()
		go func() { // read from handler(blocks) and write to peer
			defer func() {
				log.Debug(tag, "writing to peer done")
				done <- struct{}{}
			}()
			for d := range tr.Transmit() {
				if d.Error != nil {
					log.Error(tag, fmt.Errorf("reading from handler: %s", d.Error.Error()))
					continue
				}

				err = wsc.WriteMessage(websocket.TextMessage, d.Bytes)
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) { // it's ok
					return
				} else if err != nil {
					log.Error(tag, fmt.Errorf("writing to peer %s", err.Error()))
					return
				}
			}
		}()

		go func() { // read from peer (blocks) and write to handler
			defer func() {
				log.Debug(tag, "reading from peer done")
				done <- struct{}{}
			}()
			for {
				_, b, err := wsc.ReadMessage()
				log.Info("read from peer", "")
				if websocket.IsCloseError(err, websocket.CloseNormalClosure) { // it's ok
					return
				} else if err != nil { // unexpected error
					log.Error(tag, fmt.Errorf("reading from peer %s", err.Error()))
					return
				}

				if err := tr.Receive(b); err != nil {
					log.Error(tag, fmt.Errorf("writing to handler %s", err.Error()))
				}
			}

		}()

		log.Info(tag, "connection ws://%s%s established", req.RemoteAddr, req.URL.String())
		<-done
		if err := tr.Close(); err != nil {
			log.Error(tag, fmt.Errorf("closing handler %s", err.Error()))
		}

		wsc.Close()
		s.Lock()
		delete(s.connections, req.RemoteAddr)
		s.Unlock()
		log.Info(tag, "connection ws://%s%s closed", req.RemoteAddr, req.URL.String())

	}
}

// Close send close message to all clients in order to let them know that
// server is going to be shutdown.
func (s *Server) Close() error {
	//s.Lock()
	s.done <- struct{}{}
	//for _, c := range s.connections {
	//	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	//	err := c.WriteMessage(websocket.CloseMessage, msg)
	//	if err != nil {
	//		log.Error("es.server", err)
	//		continue
	//	}
	//}
	//s.Unlock()

	ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
	return s.http.Shutdown(ctx)
}

func (s *Server) Start() error {
	m := http.NewServeMux()
	m.HandleFunc("/stream", s.serve(newStreamer))
	m.HandleFunc("/subscribe", s.serve(newSubscriber))

	s.http = &http.Server{Addr: s.opts.address, Handler: m}

	log.Info("es.server", "listening on http://%s", s.opts.address)
	if err := s.http.ListenAndServe(); err != nil {
		log.Info("es.server", "exit %s", err.Error())
	}

	return nil
}

func New(opts ...Option) *Server {
	return &Server{
		connections: make(map[string]*websocket.Conn),
		opts:        newOptions(opts...),
		done:        make(chan struct{}),
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

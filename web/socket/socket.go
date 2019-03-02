package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Logger interface{ Print(typ, format string, args ...interface{}) }

type Options struct {
	Logger   Logger
	Shutdown context.Context
}

func Connect(url string, r Handler, o *Options) error {
	go func() {
		defer log.Print("DBG", "client finished")

		for { //reconnection loop
			var (
				err    error
				res    *http.Response
				socket *websocket.Conn
			)

			// connect to websocket server
			if socket, res, err = websocket.DefaultDialer.Dial(url, nil); err != nil {
				if res != nil && res.Body != nil {
					b, _ := ioutil.ReadAll(res.Body)
					err = fmt.Errorf("%s %s", res.Status, string(b))
				}

				log.Print("DBG", "connection failed due %s [%T]", err, err)

				time.Sleep(time.Second * 2)
				continue
			}
			//shutdown = context.TODO()
			server := NewPeer(socket, o.Logger)
			connection := NewConnection(server, server.alive)

			go r.ServeWS(connection)

			select {
			case <-server.alive.Done():
				continue

			case <-o.Shutdown.Done():
				if err := server.Close(); err != nil {
					log.Print("DBG", "closing server failed due %s", err)
				}
				return
			}

		}
	}()

	return nil
}

func Serve(h Handler, opts *Options) http.HandlerFunc {
	var (
		err      error
		socket   *websocket.Conn
		upgrader = websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
	)

	return func(w http.ResponseWriter, r *http.Request) {
		if socket, err = upgrader.Upgrade(w, r, nil); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		client := NewPeer(socket, opts.Logger)
		connection := NewConnection(client, client.alive)

		go h.ServeWS(connection)

		select { // wait for disconnection from client or server shutdown
		case <-connection.Termination.Done():
		case <-opts.Shutdown.Done():
			if err := client.Close(); err != nil {
				opts.Logger.Print("dbg", "closing client failed due %s", err)
			}
		}
	}
}

func decode(what []byte, into ...interface{}) error {
	for i := range into {
		if err := json.Unmarshal(what, into[i]); err != nil {
			return err
		}
	}

	return nil
}

func encode(what ...interface{}) ([]byte, error) {
	var d map[string]interface{}

	for i := range what {
		if what[i] == nil {
			continue
		}

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

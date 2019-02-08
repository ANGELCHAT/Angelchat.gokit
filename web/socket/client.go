package socket

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	connection string
	url        string
	log        logger
	shutdown   context.Context
	writer     chan []byte

	//wb bytes.Buffer
	//rb bytes.Buffer
}

func NewClient(url string, l logger) *Client {
	termination, cancel := context.WithCancel(context.Background())
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop,
		syscall.SIGTERM,
		os.Interrupt,
		os.Kill,
	)

	go func() {

		fmt.Printf("wait for shutdown...\n")

		sig := <-gracefulStop
		cancel()

		fmt.Printf("caught sig: %+v, waiting...\n", sig)
		time.Sleep(time.Second * 5)

		os.Exit(0)
	}()

	return &Client{
		url:      url,
		log:      l,
		shutdown: termination,
		writer:   make(chan []byte),
	}

}

func (c *Client) Run(h Handler) error {
	defer c.log.Info("disconnected")

	for {
		var (
			err                   error
			res                   *http.Response
			server                *websocket.Conn
			sentinel              sync.WaitGroup
			disconnection, cancel = context.WithCancel(context.Background())
		)

		if server, res, err = websocket.DefaultDialer.Dial(c.url, nil); err != nil {
			if res != nil && res.Body != nil {
				b, _ := ioutil.ReadAll(res.Body)
				err = fmt.Errorf("%s %s", res.Status, string(b))
			}

			c.log.Debug("connection failed due %s [%T]", err, err)
			c.log.Debug("reconnecting...")

			time.Sleep(time.Second * 2)
			continue
		}
		c.log.Info("connected")

		// Time allowed to read the next pong message from the client.
		server.SetPongHandler(func(s string) error {
			//c.log.Debug("pong %s %s", s, rd)
			return nil
		})

		server.SetPingHandler(func(s string) error {
			//c.log.Debug("ping %s %s", s, rd)
			return nil
		})

		server.SetCloseHandler(func(code int, text string) error {
			c.log.Debug("close message received from %d:%s", code, text)
			cancel()
			return nil
		})

		//create and configure Request structure
		dispatch := func(b []byte) {
			var m message
			if err := decode(b, &m); err != nil {
				c.log.Debug("dispatch: decode failed due %s ", err)
				return
			}

			h.ServeWS(&Message{
				message:     m,
				Body:        b,
				Termination: disconnection,
				stream:      c.writer,
			})
		}

		read := func(socket *websocket.Conn) {
			sentinel.Add(1)
			c.log.Debug("read goroutine started")
			go func() {
				defer c.log.Debug("read goroutine finished")
				defer sentinel.Done()
				defer cancel()

				for {
					var (
						b   []byte
						err error
					)

					if _, b, err = socket.ReadMessage(); err != nil {
						if _, ok := err.(*websocket.CloseError); ok {
							c.log.Debug("read: socket closed due %s ", err.Error())
							return
						}

						c.log.Debug("read: socket failed due %s ", err)
						return
					}

					//if _, err := c.wb.Write(b); err != nil {
					//	c.log.Debug("read: writing to buffer failed due %s ", err)
					//	continue
					//}

					go dispatch(b)
				}
			}()
		}

		write := func(socket *websocket.Conn) {
			sentinel.Add(1)
			c.log.Debug("write goroutine started")
			go func() {
				ping := time.NewTicker((time.Second * 9) / 10)

				defer c.log.Debug("write goroutine finished")
				defer ping.Stop()
				defer sentinel.Done()
				defer cancel()

				for {
					select {

					case <-ping.C:
						// ping client to check if connection is still established.
						if err := socket.WriteMessage(websocket.PingMessage, nil); err != nil {
							c.log.Debug("write: ping failed due %s", err.Error())
							return
						}

					case r := <-c.writer:
						if err := socket.WriteMessage(websocket.TextMessage, r); err != nil {
							if _, ok := err.(*websocket.CloseError); ok {
								c.log.Debug("write: socket closed due %s ", err.Error())
								return
							}
							c.log.Info("write: %s", err)
							continue
						}

						//c.log.Debug("write: received %+v", r)
					}
				}
			}()
		}

		read(server)
		write(server)

		var shutdown bool

		select {
		// flux.server disconnected, continue
		case <-disconnection.Done():
			c.log.Info("disconnected from flux.server")
			break

		// flux.client received signal to shutdown
		case <-c.shutdown.Done():
			// wait for all goroutines
			c.log.Info("sending close message to flux.server")
			m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "flux.client shutting down")
			server.WriteControl(websocket.CloseMessage, m, time.Now().Add(time.Second))
			shutdown = true
			break

		}

		c.log.Debug("waiting for read and write to finish")
		sentinel.Wait()

		c.log.Debug("cleaned up after disconnection from flux.server")

		if shutdown {
			return server.Close()
		}
	}
}

//func (c *Client) Read(p []byte) (n int, err error) {
//	return c.wb.Read(p)
//}
//
//func (c *Client) Write(p []byte) (n int, err error) {
//	c.writer <- p
//	fmt.Println("zapisał")
//	return c.rb.Write(p)
//}
//
//func (c *Client) Close() error {
//	panic("implement me")
//}
//

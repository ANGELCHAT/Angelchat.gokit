package socket

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Peer struct {
	alive              context.Context
	connection         *websocket.Conn
	incoming, outgoing chan []byte
}

func NewPeer(socket *websocket.Conn, log Logger) *Peer {
	var (
		sentinel     sync.WaitGroup
		c, interrupt = context.WithCancel(context.Background())
		peer         = &Peer{
			connection: socket,
			incoming:   make(chan []byte),
			outgoing:   make(chan []byte),
			alive:      c,
		}
	)

	defer log.Print("INF", "%s connected", socket.RemoteAddr().String())

	// Time allowed to read the next pong message from the client.
	socket.SetPongHandler(func(s string) error { return nil })
	socket.SetPingHandler(func(s string) error { return nil })
	socket.SetCloseHandler(func(code int, text string) error {
		// receive close frames from client socket
		//log.Debug("close frame received from peer %d:%s", code, text)
		interrupt()
		return nil
	})

	// read bytes from websocket and writes them into outgoing channel,
	// method finish when socket is disconnected or Peer has been closed.
	read := func(s *websocket.Conn) {
		sentinel.Add(1)
		go func() {
			defer func() {
				//log.Debug("read: goroutine finished")
				interrupt()
				sentinel.Done()
			}()

			for {
				_, body, err := s.ReadMessage()
				if err != nil {
					return
				}

				select { // wait until bytes are send or peer is closed
				case peer.outgoing <- body: // bytes from socket sent to outgoing channel
				case <-peer.alive.Done(): // reading from socket interrupted
					return
				}
			}
		}()
	}

	// write take care of deliver messages to websocket client, it finished
	// when websocket client not responding or when done channel is closed.
	write := func(s *websocket.Conn) {
		sentinel.Add(1)
		ping := time.NewTicker((time.Second * 3 * 9) / 10)

		go func() {
			defer func() {
				//log.Debug("write: goroutine finished")
				interrupt()
				ping.Stop()
				sentinel.Done()
			}()

			for {
				select {
				case <-peer.alive.Done(): // finish writing to socket
					return

				case <-ping.C: // ping checks if socket is still established.
					if err := s.WriteMessage(websocket.PingMessage, nil); err != nil {
						return
					}


				case body := <-peer.incoming: // bytes from incoming channel sent to socket
					if err := s.WriteMessage(websocket.TextMessage, body); err != nil {
						return
					}

					//atomic.AddUint64(c.counter, 1)
				}
			}
		}()
	}

	coordinate := func(s *websocket.Conn) {
		go func() {
			defer func() {
				//log.Debug("coordinate: goroutine finished")
				log.Print("INF", "%s disconnected", s.RemoteAddr().String())
			}()

			select {
			//case <-shutdown.Done():
			//	log.Debug("shutdown signal received")
			//	fmt.Println(p.Close())

			case <-peer.alive.Done(): // wait for disconnection from socket
				log.Print("DBG", "interrupt signal received")
			}

			sentinel.Wait()

			// todo close them?
			//p.outgoing = nil
			//p.sender = nil
		}()

	}

	read(socket)
	write(socket)
	coordinate(socket)

	return peer
}

// Read
func (p *Peer) Read(b []byte) (n int, err error) {
	select {
	case <-p.alive.Done(): // reading from disconnected socket returns end of file error.
		return 0, io.EOF

	case src := <-p.outgoing:
		return copy(b, append(src, '\n')), nil
	}
}

// Write
func (p *Peer) Write(b []byte) (n int, err error) {
	select {
	case <-p.alive.Done(): // writing to disconnected socket returns end of file error.
		return 0, io.EOF

	case p.incoming <- b:
		return len(b), nil
	}
}

// Close
func (p *Peer) Close() error {
	if err := p.connection.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseGoingAway, "shutting down"),
		time.Now().Add(time.Second)); err != nil {
		return err
	}

	return p.connection.Close()
}

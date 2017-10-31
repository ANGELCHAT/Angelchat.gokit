package client

import (
	"fmt"

	"sync"

	"github.com/gorilla/websocket"
	"github.com/sokool/gokit/log"
)

type Subscriber struct {
	conn          *websocket.Conn
	subscriptions []Subscription
	mx            sync.Mutex
}

type message struct {
	Name          string
	Subscriptions []Subscription
}

func (r *Subscriber) Read(f func([]Event)) {
	defer func() {
		log.Info("es.client.subscriber", "disconnected")
		//r.conn.Close()

	}()

	m := message{
		Name:          "subscription",
		Subscriptions: r.subscriptions,
	}

	if err := r.conn.WriteJSON(m); err != nil {
		log.Error("es.client.reader.subscribing", err)
	}

	for {
		events := new([]Event)
		err := r.conn.ReadJSON(events)
		if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
			log.Info("es.client.subscriber", "%s", err.Error())
			return
		} else if err != nil {
			log.Error("es.client.subscriber", err)
			break
		}

		f(*events)
	}
}

func (r *Subscriber) Subscribe(s []Subscription) {
	if err := r.send(s); err != nil {
		log.Error("es.client.reader",
			fmt.Errorf("sending subscription %s", err.Error()))
	}

	r.receive(func(es []Event) {

	})

}

func (r *Subscriber) send(s []Subscription) error {
	m := message{
		Name:          "subscription",
		Subscriptions: s,
	}

	return r.conn.WriteJSON(m)

}

// receive message from peer
func (r *Subscriber) receive(f func([]Event)) error {
	for {
		m := []Event{}
		err := r.conn.ReadJSON(&m)
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) { // it's ok
			return nil
		} else if err != nil { // unexpected error
			log.Error("es.client.reader", fmt.Errorf("reading from peer %s", err.Error()))
			return err
		}

		f(m)
	}

	return nil
}

func (r *Subscriber) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "interupt")
	if err := r.conn.WriteMessage(websocket.CloseMessage, msg); err != nil {
		log.Error("es.client.reader.close", err)
		return err
	}

	return nil
}

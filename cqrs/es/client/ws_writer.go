package client

import (
	"encoding/json"
	"reflect"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sokool/gokit/log"
)

type Writer struct {
	stream string
	conn   *websocket.Conn
}

func (w *Writer) Write(streamID string, meta map[string]string, events ...interface{}) error {
	var es []event
	for _, e := range events {
		b, err := json.Marshal(e)
		if err != nil {
			return err
		}

		es = append(es, event{
			ID:   uuid.New().String(),
			Data: b,
			Name: reflect.TypeOf(e).Name(),
		})
	}
	//log.Info("es.client.writer", "after events")
	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	//log.Info("es.client.writer", "sending stream")
	s := stream{
		ID:     streamID,
		Name:   w.stream,
		Meta:   b,
		Events: es,
	}

	err = w.conn.WriteJSON(s)
	if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
		log.Info("es.client.writer", "%s", err.Error())
		return nil
	} else if err != nil {
		log.Error("es.client.writer", err)
		return err
	}
	//0,2ms
	log.Info("es.client.writer", "after send")
	_, m, err := w.conn.ReadMessage()
	if err != nil {
		w.conn.Close()
		return err
	}

	log.Info("es.client.writer", "received %s", string(m))
	return nil
}

func (w *Writer) Close() {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := w.conn.WriteMessage(websocket.CloseMessage, msg)
	if err != nil {
		log.Error("es.client.writer.close", err)
		return
	}

	w.conn.Close()
}

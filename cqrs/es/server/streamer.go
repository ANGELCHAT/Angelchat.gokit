package server

import (
	"encoding/json"

	"github.com/sokool/gokit/cqrs/es"
	"github.com/sokool/gokit/log"
)

type streamer struct {
	store    *es.Service
	response chan []byte
	closed   bool
	z int
}


func (s *streamer) Transmit() <-chan Data {
	out := make(chan Data, 1)
	go func() {
		defer close(out)
		for data := range s.response {
			//if s.z > 3 {
			//	//log.Info("zamykam się", "")
			//	break
			//}
			log.Debug("es.server.streamer", "transmit response")
			out <- Data{Bytes: data, Error: nil}
			//s.z++
		}
	}()

	return out
}

func (s *streamer) Receive(in []byte) error {
	stream := es.Stream{}
	if err := json.Unmarshal(in, &stream); err != nil {
		return err
	}

	if _, err := s.store.Append(stream); err != nil {
		return err
	}
	log.Debug("es.server.streamer", "received stream")

	s.response <- []byte("OK")

	return nil
}

func (s *streamer) Close() error {
	if s.closed {
		return nil
	}
	close(s.response)
	s.closed = true
	return nil
}

func newStreamer(e *es.Service) TransmitReceiver {
	return &streamer{
		store:    e,
		response: make(chan []byte, 1),
	}
}

package log_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/livechat/gokit/log"
)

func TestName(t *testing.T) {
	logger := log.New(os.Stdout,
		log.WithLevels(true),
		log.WithTag("[TEST]"),
		log.WithTime("2 Jan 15:04:05.000000"),
		log.WithFilename(2),
	)

	logger.Print("standardowe, nic się nie dzieje.")           // "." -> INFO
	logger.Print("a to bardziej dosadne")                      // "" -> DEBUG
	logger.Print("i tutaj może być problem %s?", "z powierza") // "?" -> WARNING
	logger.Print("coś tu jest nie tak;")                       // ";" -> EMERGENCY
	logger.Print("a to błont!")                                // "!" -> ERROR
	logger.Print("a tu jafny błont: %s", fmt.Errorf("oh no"))  // error type as argument -> ERROR
	logger.Write([]byte("testuje z writera"))

}

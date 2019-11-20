package log_test

import (
	"fmt"
	logg "log"
	"testing"

	"github.com/livechat/gokit/log"
)

func TestName(t *testing.T) {
	//arrow := func(message []byte, args ...interface{}) log.Tag {
	//	return log.Tag{"arrow", `>`}
	//}

	//custom := func(w io.Writer, tags ...log.Tag) error {
	//	for i := range tags {
	//		if tags[i].Name == "message" {
	//			fmt.Println(tags[i].Value)
	//			return nil
	//		}
	//	}
	//	return nil
	//}

	logger := log.New(
		//log.WithWriter(os.Stdo),
		//log.WithTagger(arrow),
		log.WithLevels(true),
		log.WithGlobal(true),
		log.WithTime("15:04:05.000000"),
		log.WithName("TEST"),
		log.WithFilename(2),
		//log.WithRenderer(custom),
		log.WithVerbose(true),
	)

	logger.Print("just info 12345678")
	logger.Print("standardowe, nic się nie dzieje.")           // "." -> DEBUG
	logger.Print("a to bardziej dosadne")                      // "" -> INFO
	logger.Print("i tutaj może być problem %s?", "z powierza") // "?" -> WARNING
	logger.Print("coś tu jest nie tak;")                       // ";" -> EMERGENCY
	logger.Print("a to błont!")                                // "!" -> ERROR
	logger.Print("a tu jafny błont: %s", fmt.Errorf("oh no"))  // error type as argument -> ERROR
	logger.Write([]byte("testuje z writera"))

	logg.Print("elo z go.loga YYY di[a")
	log.Print("dupa z log.std XXX.")
	//time.Sleep(time.Second)

}

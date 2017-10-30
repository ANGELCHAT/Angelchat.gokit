package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/sokool/gokit/cqrs/es/server"
)

type CreatedEvent struct {
	Name    string
	Address string
}

func main() {
	//trace.Start(os.Stdout)
	//defer trace.Stop()

	srv := server.New()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)

	go func() {
		<-stop
		srv.Close()
	}()

	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}

	//s, _ := es.NewService("/tmp/ws-test")
	//ez, err := s.Events("1312", 1020200)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//for _, e := range ez {
	//	fmt.Println(e.ID, e.Version, e.Created, e.Type)
	//	if e.Type == "CreatedEvent" {
	//		z := CreatedEvent{}
	//		json.Unmarshal(e.Data, &z)
	//		fmt.Printf("%+v\n", z)
	//	}
	//}
}

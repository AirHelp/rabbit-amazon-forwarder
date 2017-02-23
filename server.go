package main

import (
	"log"
	"net/http"

	"github.com/AirHelp/rabbit-amazon-forwarder/mapping"
	"github.com/AirHelp/rabbit-amazon-forwarder/supervisor"
)

func main() {
	consumerForwarderMap, err := mapping.New().Load()
	if err != nil {
		log.Fatalf("Could not load consumer->forwader pairs. Error: " + err.Error())
	}
	supervisor := supervisor.New(consumerForwarderMap)
	if err := supervisor.Start(); err != nil {
		log.Fatal("Could not start supervisor. Error: ", err.Error())
	}
	http.HandleFunc("/restart", supervisor.Restart)
	http.HandleFunc("/health", supervisor.Check)
	log.Print("Starting http server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

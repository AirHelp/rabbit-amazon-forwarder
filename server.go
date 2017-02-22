package main

import (
	"log"
	"net/http"

	"github.com/AirHelp/rabbit-amazon-forwarder/healthcheck"
	"github.com/AirHelp/rabbit-amazon-forwarder/mapping"
)

func main() {
	http.HandleFunc("/health", health.Check)
	err := mapping.New().LoadAndStart()
	if err != nil {
		log.Fatalf("Could not load and start consumer->forwader pairs. Error: " + err.Error())
	}
	log.Print("Starting http server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

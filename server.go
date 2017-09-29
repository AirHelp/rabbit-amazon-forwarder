package main

import (
	"github.com/AirHelp/rabbit-amazon-forwarder/mapping"
	"github.com/AirHelp/rabbit-amazon-forwarder/supervisor"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func main() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)

	consumerForwarderMap, err := mapping.New().Load()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error()}).Fatalf("Could not load consumer - forwarder pairs. Error: " + err.Error())
	}
	supervisor := supervisor.New(consumerForwarderMap)
	if err := supervisor.Start(); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error()}).Fatal("Could not start supervisor")
	}
	http.HandleFunc("/restart", supervisor.Restart)
	http.HandleFunc("/health", supervisor.Check)
	log.Info("Starting http server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"net/http"
	"os"

	"github.com/AirHelp/rabbit-amazon-forwarder/mapping"
	"github.com/AirHelp/rabbit-amazon-forwarder/supervisor"
	log "github.com/sirupsen/logrus"
)

const (
	LogLevel = "LOG_LEVEL"
)

func main() {
	createLogger()

	consumerForwarderMap, err := mapping.New().Load()
	if err != nil {
		log.WithField("error", err.Error()).Fatalf("Could not load consumer - forwarder pairs")
	}
	supervisor := supervisor.New(consumerForwarderMap)
	if err := supervisor.Start(); err != nil {
		log.WithField("error", err.Error()).Fatal("Could not start supervisor")
	}
	http.HandleFunc("/restart", supervisor.Restart)
	http.HandleFunc("/health", supervisor.Check)
	http.HandleFunc("/stop", supervisor.Stop)

	log.Info("Starting http server")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func createLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	if logLevel := os.Getenv(LogLevel); logLevel != "" {
		if level, err := log.ParseLevel(logLevel); err != nil {
			log.Fatal(err)
		} else {
			log.SetLevel(level)
		}
	}
}

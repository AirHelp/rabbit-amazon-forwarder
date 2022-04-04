package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/symopsio/rabbit-amazon-forwarder/mapping"
	"github.com/symopsio/rabbit-amazon-forwarder/supervisor"
	"net/http"
	"os"
)

const (
	LogLevel = "LOG_LEVEL"
)

func main() {
	createLogger()

	consumerForwarderMapping, err := mapping.New().Load()
	if err != nil {
		log.WithField("error", err.Error()).Fatalf("Could not load consumer - forwarder pairs")
	}
	supervisor := supervisor.New(consumerForwarderMapping)
	if err := supervisor.Start(); err != nil {
		log.WithField("error", err.Error()).Fatal("Could not start supervisor")
	}
	http.HandleFunc("/restart", supervisor.Restart)
	http.HandleFunc("/health", supervisor.Check)
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

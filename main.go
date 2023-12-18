package main

import (
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/kafka"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	appConfig, err := config.LoadConfig("app.yml")
	if err != nil {
		log.WithError(err).Fatal("Error loading config file")
		os.Exit(1)
	}
	configureLogger(appConfig.App)

	p := kafka.NewProcessor(*appConfig)
	p.Run()
}

func configureLogger(config config.App) {
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)
}

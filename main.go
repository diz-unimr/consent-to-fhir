package main

import (
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/kafka"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	appConfig := loadConfig()
	configureLogger(appConfig.App)

	p := kafka.NewProcessor(appConfig)
	p.Run()
}

func loadConfig() config.AppConfig {
	c, err := config.LoadConfig(".")
	if err != nil {
		log.WithError(err).Fatal("Unable to load config file")
	}
	return c
}

func configureLogger(config config.App) {
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)
}

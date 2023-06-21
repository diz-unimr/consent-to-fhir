package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type AppConfig struct {
	App   App   `mapstructure:"app"`
	Kafka Kafka `mapstructure:"kafka"`
	Gics  Gics  `mapstructure:"gics"`
}

type App struct {
	Name     string `mapstructure:"name"`
	LogLevel string `mapstructure:"log-level"`
	Mapper   Mapper `mapstructure:"mapper"`
}

type Mapper struct {
	ConsentSystem *string `mapstructure:"consent-system"`
	PatientSystem *string `mapstructure:"patient-system"`
}

type Kafka struct {
	BootstrapServers string `mapstructure:"bootstrap-servers"`
	InputTopic       string `mapstructure:"input-topic"`
	OutputTopic      string `mapstructure:"output-topic"`
	SecurityProtocol string `mapstructure:"security-protocol"`
	Ssl              Ssl    `mapstructure:"ssl"`
	NumConsumers     int    `mapstructure:"num-consumers"`
}

type Gics struct {
	SignerId string `mapstructure:"signer-id"`
	Fhir     Fhir   `mapstructure:"fhir"`
}

type Ssl struct {
	CaLocation          string `mapstructure:"ca-location"`
	CertificateLocation string `mapstructure:"certificate-location"`
	KeyLocation         string `mapstructure:"key-location"`
	KeyPassword         string `mapstructure:"key-password"`
}

type Fhir struct {
	Base string `mapstructure:"base"`
	Auth *Auth  `mapstructure:"auth"`
}

type Auth struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

func LoadConfig() AppConfig {
	c, err := parseConfig(".")
	if err != nil {
		log.WithError(err).Fatal("Unable to load config file")
		os.Exit(1)
	}

	return *c
}

func parseConfig(path string) (config *AppConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("yml")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`, `-`, `_`))

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	return
}

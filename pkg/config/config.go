package config

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"regexp"
	"strings"
)

type AppConfig struct {
	App   App   `koanf:"app"`
	Kafka Kafka `koanf:"kafka"`
	Gics  Gics  `koanf:"gics"`
}

type App struct {
	Name     string `koanf:"name"`
	LogLevel string `koanf:"log-level"`
	Mapper   Mapper `koanf:"mapper"`
}

type Mapper struct {
	ConsentSystem *string           `koanf:"consent-system"`
	PatientSystem *string           `koanf:"patient-system"`
	DomainSystem  *string           `koanf:"domain-system"`
	Profiles      map[string]string `koanf:"profiles"`
}

type Kafka struct {
	BootstrapServers string `koanf:"bootstrap-servers"`
	InputTopic       string `koanf:"input-topic"`
	OutputTopic      string `koanf:"output-topic"`
	SecurityProtocol string `koanf:"security-protocol"`
	Ssl              Ssl    `koanf:"ssl"`
	NumConsumers     int    `koanf:"num-consumers"`
}

type Gics struct {
	Fhir Fhir `koanf:"fhir"`
}

type Ssl struct {
	CaLocation          string `koanf:"ca-location"`
	CertificateLocation string `koanf:"certificate-location"`
	KeyLocation         string `koanf:"key-location"`
	KeyPassword         string `koanf:"key-password"`
}

type Fhir struct {
	Base string `koanf:"base"`
	Auth *Auth  `koanf:"auth"`
}

type Auth struct {
	User     string `koanf:"user"`
	Password string `koanf:"password"`
}

func LoadConfig(path string) (*AppConfig, error) {

	// load config file
	var k = koanf.New(".")
	f := file.Provider(path)
	if err := k.Load(f, yaml.Parser()); err != nil {
		return nil, err
	}
	// replace env vars
	_ = k.Load(env.Provider("", ".", func(s string) string {
		return parseEnv(k, s)
	}), nil)

	return parseConfig(k), nil
}

func parseEnv(k *koanf.Koanf, s string) string {
	r := "^" + strings.Replace(strings.ToLower(s), "_", "(.|-)", -1) + "$"

	for _, p := range k.Keys() {
		match, _ := regexp.MatchString(r, p)
		if match {
			return p
		}
	}
	return ""
}

func parseConfig(k *koanf.Koanf) (config *AppConfig) {
	_ = k.Unmarshal("", &config)
	return config
}

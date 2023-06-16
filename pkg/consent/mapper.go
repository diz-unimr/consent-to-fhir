package consent

import (
	"consent-to-fhir/pkg/client"
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/model"
	"encoding/json"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"
)

func NewConsentMapper(c config.AppConfig) *Mapper {
	return &Mapper{
		Client: client.NewGicsClient(c),
	}
}

type Mapper struct {
	Client *client.GicsClient
}

func (m *Mapper) Process(data []byte) *fhir.Bundle {
	var n model.Notification
	err := json.Unmarshal(data, &n)
	if err != nil {
		log.WithError(err)
		return nil
	}

	bundle := m.toFhir(n)

	return bundle
}

func (m *Mapper) toFhir(n model.Notification) *fhir.Bundle {
	// get current consent state from gics
	bundle, err := m.Client.GetConsentStatus(n.ConsentKey.SignerIds[0], *n.ConsentKey.ConsentTemplateKey)
	if err != nil {
		log.Error("Request to get consent status from gICS failed")
		return nil
	}
	return bundle
}

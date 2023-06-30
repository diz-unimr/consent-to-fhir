package mapper

import (
	"consent-to-fhir/pkg/config"
	"fmt"
	"github.com/magiconair/properties/assert"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"io"
	"os"
	"testing"
)

func TestNewGicsMapper(t *testing.T) {
	c := config.AppConfig{
		App:   config.App{},
		Kafka: config.Kafka{},
		Gics: config.Gics{Fhir: config.Fhir{
			Base: "base",
			Auth: &config.Auth{
				User:     "foo",
				Password: "bar",
			}}},
	}

	m := NewGicsMapper(c)

	assert.Equal(t, m.Client.GetRequestUrl(), "base/$currentPolicyStatesForPerson")
	assert.Equal(t, m.Client.GetAuth(), c.Gics.Fhir.Auth)
	assert.Equal(t, m.Config, c.App.Mapper)
	assert.Equal(t, m.ConsentProfile, NewConsentProfile(MiiProfile))
}

type TestGicsClient struct{}

func (c *TestGicsClient) GetRequestUrl() string {
	return ""
}

func (c *TestGicsClient) GetAuth() *config.Auth {
	return nil
}

func (c *TestGicsClient) RequestUrl() string {
	return ""
}

func (c *TestGicsClient) Auth() *config.Auth {
	return nil
}

func (c *TestGicsClient) GetConsentStatus(signerId string, domain string, date string) (*fhir.Bundle, error) {
	testFile, _ := os.Open("testdata/current-policies-response.json")
	b, _ := io.ReadAll(testFile)

	bundle, err := fhir.UnmarshalBundle(b)

	return &bundle, err
}

func TestProcess(t *testing.T) {
	c := config.AppConfig{App: config.App{Mapper: config.Mapper{
		ConsentSystem: Of("https://fhir.diz.uni-marburg.de/sid/consent-id"),
		PatientSystem: Of("https://fhir.diz.uni-marburg.de/sid/patient-id"),
		DomainSystem:  Of("https://fhir.diz.uni-marburg.de/fhir/sid/consent-domain-id"),
	}}}

	m := NewGicsMapper(c)
	m.Client = &TestGicsClient{}
	input := []byte(`
		{
		  "consentKey": {
			"consentTemplateKey": {
			  "domainName": "MII",
			  "name": "Patienteneinwilligung MII",
			  "version": "1.6.d"
			},
			"signerIds": [
			  {
				"idType": "Patienten-ID",
				"id": "42",
				"orderNumber": 0
			  }
			],
			"consentDate": "2023-05-02 01:57:27"
		  }
		}
	`)

	expected := fhir.Consent{
		Meta: &fhir.Meta{
			Profile: []string{MiiProfile},
		},
		Patient: &fhir.Reference{
			Reference: Of(fmt.Sprintf("Patient?identifier=%s|%s", *c.App.Mapper.PatientSystem, "42")),
		},
		Policy: []fhir.ConsentPolicy{{Uri: m.ConsentProfile.ConsentPolicy.Uri}},
	}
	bundle := m.Process(input)
	actual, _ := fhir.UnmarshalConsent(bundle.Entry[0].Resource)

	assert.Equal(t, actual.Meta.Profile, expected.Meta.Profile)
	assert.Equal(t, *actual.Patient.Reference, *expected.Patient.Reference)
	assert.Equal(t, actual.Policy, expected.Policy)
}

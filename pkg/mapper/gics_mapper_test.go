package mapper

import (
	"consent-to-fhir/pkg/config"
	"github.com/magiconair/properties/assert"
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

	assert.Equal(t, m.Client.RequestUrl, "base/$currentPolicyStatesForPerson")
	assert.Equal(t, m.Client.Auth, c.Gics.Fhir.Auth)
	assert.Equal(t, m.Config, c.App.Mapper)
	assert.Equal(t, m.ConsentProfile, NewConsentProfile(MiiProfile))
}

func TestGetPolicyUri(t *testing.T) {

}

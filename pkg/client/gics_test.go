package client

import (
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/model"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetConsentDomain(t *testing.T) {

	id := "test-id"
	study := fhir.ResearchStudy{Id: &id}
	res, _ := study.MarshalJSON()

	s := withTestServer(res, 200)
	defer s.Close()

	c := NewGicsClient(config.AppConfig{Gics: config.Gics{
		Fhir: config.Fhir{Base: s.URL},
	}})

	actual, _ := c.GetConsentDomain("/ResearchStudy/" + id)

	assert.Equal(t, id, *actual.Id)
}

func TestGetConsentStatus(t *testing.T) {

	testId := "test"
	res, _ := fhir.Consent{Id: &testId}.MarshalJSON()
	b, _ := fhir.Bundle{
		Entry: []fhir.BundleEntry{{Resource: res}},
	}.MarshalJSON()

	s := withTestServer(b, 200)
	defer s.Close()

	c := NewGicsClient(config.AppConfig{Gics: config.Gics{
		Fhir: config.Fhir{Base: s.URL},
	}})

	actual, _ := c.GetConsentStatus(model.SignerId{Id: "test"}, "domain", "2024-01-01")
	resource, _ := actual.Entry[0].Resource.MarshalJSON()
	consent, _ := fhir.UnmarshalConsent(resource)

	assert.Equal(t, testId, *consent.Id)
}

func withTestServer(response []byte, code int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		res.WriteHeader(code)
		_, _ = res.Write(response)
	}))
}

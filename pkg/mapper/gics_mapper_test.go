package mapper

import (
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/model"
	"fmt"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

type MergePolicyTestCase struct {
	name      string
	resources []fhir.Consent
	expected  []fhir.Coding
}

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

func (c *TestGicsClient) GetConsentStatus(_ model.SignerId, _, _ string) (*fhir.Bundle, error) {
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
		Profiles: map[string]string{
			"MII": "https://www.medizininformatik-initiative.de/fhir/modul-consent/StructureDefinition/mii-pr-consent-einwilligung",
		},
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
		Policy: []fhir.ConsentPolicy{
			{
				// Patienteneinwilligung MII|1.6.d
				Uri: Of("urn:oid:2.16.840.1.113883.3.1937.777.24.2.1790"),
			},
		},
	}
	bundle := m.Process(input)
	actual, _ := fhir.UnmarshalConsent(bundle.Entry[0].Resource)

	assert.Equal(t, actual.Meta.Profile, expected.Meta.Profile)
	assert.Equal(t, *actual.Patient.Reference, *expected.Patient.Reference)
	assert.Equal(t, actual.Policy, expected.Policy)
}

func TestMergePolicies(t *testing.T) {

	discFoo := fhir.Coding{System: Of("https://system-id.local/test-policy-discarded"), Code: Of("Discarded-Foo")}
	expFoo := fhir.Coding{System: Of("https://system-id.local/test-policy-expected"), Code: Of("Expected-Foo")}
	discBar := fhir.Coding{System: Of("https://system-id.local/test-policy-discarded"), Code: Of("Discarded-Bar")}
	expBar := fhir.Coding{System: Of("https://system-id.local/test-policy-expected"), Code: Of("Expected-Bar")}

	cases := []MergePolicyTestCase{
		{"provisionsAreMerged", []fhir.Consent{
			createConsentTestData([]fhir.Coding{discFoo, expFoo}, fhir.ConsentProvisionTypePermit),
			createConsentTestData([]fhir.Coding{discBar, expBar}, fhir.ConsentProvisionTypePermit),
		}, []fhir.Coding{expFoo, expBar}},

		{"typesAreFiltered", []fhir.Consent{
			createConsentTestData([]fhir.Coding{discFoo, expFoo}, fhir.ConsentProvisionTypePermit),
			createConsentTestData([]fhir.Coding{discBar, expBar}, fhir.ConsentProvisionTypeDeny),
		}, []fhir.Coding{expFoo}},
		{"codingsAreFiltered", []fhir.Consent{
			createConsentTestData([]fhir.Coding{}, fhir.ConsentProvisionTypePermit),
			createConsentTestData([]fhir.Coding{discBar, expBar}, fhir.ConsentProvisionTypePermit),
		}, []fhir.Coding{expBar}}}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			// arrange input resources
			var entries []fhir.BundleEntry
			for _, r := range c.resources {
				// one consent resource per provision (policy)
				res, _ := r.MarshalJSON()
				entries = append(entries, fhir.BundleEntry{Resource: res})
			}

			p := mergePolicies(entries)
			var actual []fhir.Coding
			for _, prov := range p {
				for _, cc := range prov.Code {
					actual = append(actual, cc.Coding...)
				}
			}

			assert.Equal(t, actual, c.expected)
		})
	}
}

func TestMergePolicies_NoExpiry(t *testing.T) {

	cases := []struct {
		name      string
		period    fhir.Period
		periodEnd *string
	}{
		{
			name:      "ok",
			period:    fhir.Period{Start: Of("2023-12-11T00:00:00+01:00"), End: Of("2053-12-11T00:00:00+01:00")},
			periodEnd: Of("2053-12-11T00:00:00+01:00"),
		},
		{
			name:      "fix",
			period:    fhir.Period{Start: Of("2023-12-11T00:00:00+01:00"), End: Of("3000-01-01T00:00:00+01:00")},
			periodEnd: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			actual := fixNoExpiryDate(&c.period)

			assert.Equal(t, *actual.Start, *c.period.Start)
			assert.Equal(t, actual.End, (*string)(c.periodEnd))
		})
	}
}

func createConsentTestData(codings []fhir.Coding, pt fhir.ConsentProvisionType) fhir.Consent {

	return fhir.Consent{
		Provision: Of(fhir.ConsentProvision{
			Type: Of(fhir.ConsentProvisionTypeDeny),
			Provision: []fhir.ConsentProvision{
				{
					Type: Of(pt),
					Code: []fhir.CodeableConcept{{Coding: codings}},
				},
			},
		})}
}

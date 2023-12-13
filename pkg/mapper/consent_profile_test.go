package mapper

import (
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestCase struct {
	name     string
	profile  string
	category []fhir.CodeableConcept
	policy   []fhir.ConsentPolicy
}

func TestGetPolicy(t *testing.T) {
	miiPolicy := []fhir.ConsentPolicy{
		{
			// Patienteneinwilligung MII|1.6.d
			Uri: Of("urn:oid:2.16.840.1.113883.3.1937.777.24.2.1790"),
		},
	}
	miiCat := []fhir.CodeableConcept{
		{
			Coding: []fhir.Coding{{System: Of("http://loinc.org"), Code: Of("57016-8")}},
		},
		{
			Coding: []fhir.Coding{{
				System: Of("https://www.medizininformatik-initiative.de/fhir/modul-consent/CodeSystem/mii-cs-consent-consent_category"),
				Code:   Of("2.16.840.1.113883.3.1937.777.24.2.184")}},
		},
	}

	cases := []TestCase{
		{"miiProfile", MiiProfile, miiCat, miiPolicy},
		{"defaultProfile", "http://fhir.de/ConsentManagement/StructureDefinition/Consent", nil, nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			runMapProfile(t, c)
		})
	}
}

func runMapProfile(t *testing.T, expected TestCase) {
	c := fhir.Consent{Meta: &fhir.Meta{Profile: []string{expected.profile}}}
	actual := MapProfile(c)

	assert.Equal(t, expected.category, actual.Category)
	assert.Equal(t, expected.policy, actual.Policy)
}

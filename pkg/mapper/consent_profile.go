package mapper

import "github.com/samply/golang-fhir-models/fhir-models/fhir"

const (
	MiiProfile          = "https://www.medizininformatik-initiative.de/fhir/modul-consent/StructureDefinition/mii-pr-consent-einwilligung"
	LoincCategorySystem = "http://loinc.org"
	LoincCategoryCode   = "57016-8"
	MiiCategorySystem   = "https://www.medizininformatik-initiative.de/fhir/modul-consent/CodeSystem/mii-cs-consent-consent_category"
	MiiCategoryCode     = "2.16.840.1.113883.3.1937.777.24.2.184"
	MiiProvisionCode    = "urn:oid:2.16.840.1.113883.3.1937.777.24.5.3"
)

type ConsentProfile struct {
	Category []fhir.CodeableConcept
}

func NewConsentProfile(profile string) *ConsentProfile {
	if profile == MiiProfile {
		return &ConsentProfile{
			Category: []fhir.CodeableConcept{
				{
					Coding: []fhir.Coding{{System: Of(LoincCategorySystem), Code: Of(LoincCategoryCode)}},
				},
				{
					Coding: []fhir.Coding{{System: Of(MiiCategorySystem), Code: Of(MiiCategoryCode)}},
				},
			},
		}
	}

	return nil
}

func GetPolicyUri() func(string) string {
	policies := map[string]string{
		"Patienteneinwilligung MII|1.6.d": "urn:oid:2.16.840.1.113883.3.1937.777.24.2.184",
	}

	return func(key string) string {
		return policies[key]
	}
}

func Of[E any](e E) *E {
	return &e
}

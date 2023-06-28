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

type ConsentPolicy struct {
	Uri  *string
	Name *string
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

	// only mii profile, currently
	return nil
}

func GetPolicy() func(string) *ConsentPolicy {
	miiPolicy := &ConsentPolicy{
		Uri:  Of("urn:oid:2.16.840.1.113883.3.1937.777.24.2.1790"),
		Name: Of("Patienteneinwilligung MII|1.6.d"),
	}

	policies := map[string]*ConsentPolicy{
		"Patienteneinwilligung MII|1.6.d":                                       miiPolicy,
		"Teilwiderruf (kompatibel zu Patienteneinwilligung MII 1.6d)|2.0.a":     miiPolicy,
		"Vollständiger Widerruf (kompatibel zu Patienteneinwilligung MII 1.6d)": miiPolicy,
	}

	return func(key string) *ConsentPolicy {
		return policies[key]
	}
}

func Of[E any](e E) *E {
	return &e
}

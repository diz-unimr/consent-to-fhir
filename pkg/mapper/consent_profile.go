package mapper

import (
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
)

const (
	MiiProfile = "https://www.medizininformatik-initiative.de/fhir/modul-consent/StructureDefinition/mii-pr-consent-einwilligung"
)

func MapProfile(consent fhir.Consent) fhir.Consent {
	if consent.Meta == nil || len(consent.Meta.Profile) == 0 {
		return consent
	}

	p := consent.Meta.Profile[0]

	// only mii profile, currently
	if p == MiiProfile {
		consent.Category = []fhir.CodeableConcept{
			{
				Coding: []fhir.Coding{{System: Of("http://loinc.org"), Code: Of("57016-8")}},
			},
			{
				Coding: []fhir.Coding{{
					System: Of("https://www.medizininformatik-initiative.de/fhir/modul-consent/CodeSystem/mii-cs-consent-consent_category"),
					Code:   Of("2.16.840.1.113883.3.1937.777.24.2.184")}},
			},
		}
		consent.Policy = []fhir.ConsentPolicy{
			{
				// Patienteneinwilligung MII|1.6.d
				Uri: Of("urn:oid:2.16.840.1.113883.3.1937.777.24.2.1790"),
			},
		}
	}

	return consent
}

//func InitProfiles() map[string]Profile {
//
//	bc := Profile{
//		Category: []fhir.CodeableConcept{
//			{
//				Coding: []fhir.Coding{{System: Of(LoincCategorySystem), Code: Of(LoincCategoryCode)}},
//			},
//			{
//				Coding: []fhir.Coding{{
//					System: Of("https://www.medizininformatik-initiative.de/fhir/modul-consent/CodeSystem/mii-cs-consent-consent_category"),
//					Code:   Of("2.16.840.1.113883.3.1937.777.24.2.184")}},
//			},
//		},
//		ConsentPolicy: &ConsentPolicy{
//			Uri:  Of("urn:oid:2.16.840.1.113883.3.1937.777.24.2.1790"),
//			Name: Of("Patienteneinwilligung MII|1.6.d"),
//		},
//	}
//
//	// only mii profile, currently
//	return map[string]Profile{
//		MiiProfile: bc,
//	}
//}

//func (p *Profile) GetPolicy(templateName string) *ConsentPolicy {
//	if p.PolicyMatcher.MatchString(templateName) {
//		return p.ConsentPolicy
//	}
//
//	log.WithField("template", templateName).Error("Unable to determine Consent policy")
//	return nil
//}

func Of[E any](e E) *E {
	return &e
}

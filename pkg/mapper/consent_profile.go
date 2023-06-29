package mapper

import (
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"
	"regexp"
)

const (
	MiiProfile          = "https://www.medizininformatik-initiative.de/fhir/modul-consent/StructureDefinition/mii-pr-consent-einwilligung"
	LoincCategorySystem = "http://loinc.org"
	LoincCategoryCode   = "57016-8"
	MiiCategorySystem   = "https://www.medizininformatik-initiative.de/fhir/modul-consent/CodeSystem/mii-cs-consent-consent_category"
	MiiCategoryCode     = "2.16.840.1.113883.3.1937.777.24.2.184"
	MiiProvisionCode    = "urn:oid:2.16.840.1.113883.3.1937.777.24.5.3"
)

type ConsentProfile struct {
	Category      []fhir.CodeableConcept
	ConsentPolicy *ConsentPolicy
	PolicyMatcher *regexp.Regexp
}

type ConsentPolicy struct {
	Uri  *string
	Name *string
}

func NewConsentProfile(profile string) *ConsentProfile {
	if profile == MiiProfile {
		regex, _ := regexp.Compile(`^.*(Patienteneinwilligung MII)[|\s](1.6d|1.6.d).*$`)
		return &ConsentProfile{
			Category: []fhir.CodeableConcept{
				{
					Coding: []fhir.Coding{{System: Of(LoincCategorySystem), Code: Of(LoincCategoryCode)}},
				},
				{
					Coding: []fhir.Coding{{System: Of(MiiCategorySystem), Code: Of(MiiCategoryCode)}},
				},
			},
			ConsentPolicy: &ConsentPolicy{
				Uri:  Of("urn:oid:2.16.840.1.113883.3.1937.777.24.2.1790"),
				Name: Of("Patienteneinwilligung MII|1.6.d"),
			},
			PolicyMatcher: regex,
		}
	}

	// only mii profile, currently
	return nil
}

func (p *ConsentProfile) GetPolicy(templateName string) *ConsentPolicy {
	if p.PolicyMatcher.MatchString(templateName) {
		return p.ConsentPolicy
	}

	log.WithField("template", templateName).Error("Unable to determine Consent policy")
	return nil
}

func Of[E any](e E) *E {
	return &e
}

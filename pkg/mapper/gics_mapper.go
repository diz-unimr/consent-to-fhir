package mapper

import (
	"consent-to-fhir/pkg/client"
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/model"
	"crypto"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"
	"strings"
)

type GicsMapper struct {
	Client         *client.GicsClient
	Config         config.Mapper
	ConsentProfile *ConsentProfile
}

func NewGicsMapper(c config.AppConfig) *GicsMapper {

	return &GicsMapper{
		Client:         client.NewGicsClient(c),
		Config:         c.App.Mapper,
		ConsentProfile: NewConsentProfile(MiiProfile),
	}
}

func (m *GicsMapper) Process(data []byte) *fhir.Bundle {
	var n model.Notification
	err := json.Unmarshal(data, &n)
	if err != nil {
		log.WithError(err)
		return nil
	}

	bundle, err := m.toFhir(n)
	if err != nil {
		log.WithError(err).Error("Failed to map consent")
		return nil
	}

	return bundle
}

func (m *GicsMapper) toFhir(n model.Notification) (*fhir.Bundle, error) {

	// get current consent state from gics
	signerId := n.ConsentKey.SignerIds[0]
	bundle, err := m.Client.GetConsentStatus(
		signerId.Id,
		*n.ConsentKey.ConsentTemplateKey.DomainName,
		*n.ConsentKey.ConsentDate,
	)
	if err != nil {
		log.Error("Request to get consent status from gICS failed")
		return nil, err
	}

	// map resources
	policyName := fmt.Sprintf(
		"%s|%s",
		*n.ConsentKey.ConsentTemplateKey.Name,
		*n.ConsentKey.ConsentTemplateKey.Version,
	)
	return m.mapResources(bundle, n.ConsentKey.ConsentTemplateKey.DomainName, signerId.Id, policyName)
}

func (m *GicsMapper) mapResources(bundle *fhir.Bundle, domain *string, pid string, policyName string) (*fhir.Bundle, error) {

	// check bundle
	if len(bundle.Entry) == 0 {
		return nil, errors.New("no Consent resource found in gICS FHIR bundle")
	}

	// prepare consent resource & merge policies
	c, _ := fhir.UnmarshalConsent(bundle.Entry[0].Resource)
	c.Provision = &fhir.ConsentProvision{
		Type:      Of(fhir.ConsentProvisionTypeDeny),
		Period:    c.Provision.Period,
		Provision: mergePolicies(bundle.Entry),
	}

	// map
	r := m.mapConsent(c, domain, pid, policyName)
	data, err := r.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// build Bundle
	return &fhir.Bundle{
		Type: fhir.BundleTypeTransaction,
		Entry: []fhir.BundleEntry{
			{
				Resource: data,
				Request: &fhir.BundleEntryRequest{
					Method: fhir.HTTPVerbPUT,
					Url:    fmt.Sprintf("Consent?identifier=%s|%s", *m.Config.ConsentSystem, *r.Id),
				}}}}, nil

}

func (m *GicsMapper) mapConsent(c fhir.Consent, domain *string, pid string, policyName string) fhir.Consent {
	// set id
	id := hash(*domain, pid)
	c.Id = &id

	// set profile
	c.Meta.Profile = []string{MiiProfile}

	// map category
	c.Category = m.ConsentProfile.Category

	// set identifier
	c.Identifier = []fhir.Identifier{{
		System: m.Config.ConsentSystem,
		Value:  &id,
	}}

	// consent policy
	policy := GetPolicy()(policyName)
	c.Policy = []fhir.ConsentPolicy{{Uri: policy.Uri}}
	// remove policyRule and source reference
	c.PolicyRule = nil
	c.SourceReference = nil

	// set patient
	p := fmt.Sprintf("Patient?identifier=%s|%s", *m.Config.PatientSystem, pid)
	c.Patient = &fhir.Reference{
		Reference: &p,
	}

	// set domain extension
	c.Extension = m.setDomainExtension(c.Extension, domain)

	return c
}

func mergePolicies(entries []fhir.BundleEntry) []fhir.ConsentProvision {
	var p []fhir.ConsentProvision

	for _, e := range entries {
		c, _ := fhir.UnmarshalConsent(e.Resource)

		// there is only single mii code per provision
		miiCode := getSingleMiiCode(c.Provision.Code)
		if miiCode == nil {
			log.WithError(errors.New("no MII coding found")).
				Warn("Failed to map provision. Consent resource might be incomplete")
			continue
		}

		c.Provision.Code = []fhir.CodeableConcept{*miiCode}
		p = append(p, *c.Provision)
	}

	return p
}

func getSingleMiiCode(codes []fhir.CodeableConcept) *fhir.CodeableConcept {
	for _, c := range codes {
		for _, coding := range c.Coding {
			if coding.System != nil && *coding.System == MiiProvisionCode {
				coding.Display = fixDisplay(coding.Display)
				c.Coding = []fhir.Coding{coding}
				return &c
			}
		}
	}
	return nil
}

func fixDisplay(display *string) *string {
	if display != nil {
		// replace underscores for correct display values
		return Of(strings.Replace(*display, "_", " ", -1))
	}
	return display
}

func (m *GicsMapper) setDomainExtension(extensions []fhir.Extension, domain *string) []fhir.Extension {

	for _, e := range extensions {
		if e.Url == "http://fhir.de/ConsentManagement/StructureDefinition/DomainReference" {
			refIndex := -1
			for i, ext := range e.Extension {
				if ext.Url == "domain" {
					refIndex = i
					break
				}
			}

			if refIndex >= 0 {
				domainRef := fmt.Sprintf("ResearchStudy?identifier=%s|%s", *m.Config.DomainSystem, *domain)
				e.Extension[refIndex] = fhir.Extension{Url: "domain", ValueReference: &fhir.Reference{Reference: &domainRef, Display: domain}}
			}
		}
	}

	return extensions
}

func hash(values ...string) string {
	h := crypto.SHA256.New()
	for _, v := range values {
		h.Write([]byte(v))
	}
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

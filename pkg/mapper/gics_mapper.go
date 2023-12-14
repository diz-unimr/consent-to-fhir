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
	"time"
)

type GicsMapper struct {
	Client client.GicsClient
	Config config.Mapper
}

func NewGicsMapper(c config.AppConfig) *GicsMapper {

	return &GicsMapper{
		Client: client.NewGicsClient(c),
		Config: c.App.Mapper,
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
		signerId,
		*n.ConsentKey.ConsentTemplateKey.DomainName,
		*n.ConsentKey.ConsentDate,
	)
	if err != nil {
		log.Error("Request to get consent status from gICS failed")
		return nil, err
	}

	// map resources
	return m.mapResources(bundle, n.ConsentKey.ConsentTemplateKey.DomainName, signerId.Id)
}

func (m *GicsMapper) mapResources(bundle *fhir.Bundle, domain *string, pid string) (*fhir.Bundle, error) {

	// check bundle
	if len(bundle.Entry) == 0 {
		return nil, errors.New("no Consent resource found in gICS FHIR bundle")
	}

	// prepare consent resource & merge policies
	c, _ := fhir.UnmarshalConsent(bundle.Entry[0].Resource)
	c.Provision = &fhir.ConsentProvision{
		Type:      Of(fhir.ConsentProvisionTypeDeny),
		Period:    fixNoExpiryDate(c.Provision.Period),
		Provision: mergePolicies(bundle.Entry),
	}

	// map
	r := m.mapConsent(c, domain, pid)
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

func (m *GicsMapper) mapConsent(c fhir.Consent, domain *string, pid string) fhir.Consent {
	// set id
	id := hash(*domain, pid)
	c.Id = &id

	// set profile and do custom mapping
	if p, ok := m.Config.Profiles[*domain]; ok {
		c.Meta.Profile = []string{p}

		// map to profile
		c = MapProfile(c)
	}

	// set identifier
	c.Identifier = []fhir.Identifier{{
		System: m.Config.ConsentSystem,
		Value:  &id,
	}}

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

		// provisions are nested
		if prov := c.Provision; prov != nil && len(prov.Provision) > 0 {

			// get first provision
			pp := prov.Provision[0]

			if *pp.Type == fhir.ConsentProvisionTypeDeny || len(pp.Code) == 0 {
				// provision already denied by first level provision or no code exists
				continue
			}

			// fix 'no end date' of provision period
			pp.Period = fixNoExpiryDate(pp.Period)

			// pick single coding from provision.code
			coding := getSingleCoding(pp.Code)
			if coding == nil {
				// no coding found
				continue
			}

			pp.Code = []fhir.CodeableConcept{{Coding: []fhir.Coding{*coding}}}
			p = append(p, pp)
		}
	}

	return p
}

func fixNoExpiryDate(period *fhir.Period) *fhir.Period {
	if period != nil && period.End != nil {
		end := parseTime(period.End)
		noExpiry := time.Date(3000, 1, 1, 0, 0, 0, 0, time.FixedZone(end.Zone()))

		if noExpiry.Equal(*end) {
			return &fhir.Period{
				Start: period.Start,
				End:   nil,
			}
		}
	}
	return period
}

func getSingleCoding(codes []fhir.CodeableConcept) *fhir.Coding {
	// just pick the last coding of the first code as it is currently the most specific one by convention
	// this might change in the future

	coding := codes[0].Coding
	if len(coding) > 0 {
		return &coding[len(coding)-1]
	}

	return nil
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

func parseTime(dt *string) *time.Time {
	t, err := time.Parse(time.RFC3339, *dt)
	if err != nil {
		log.WithError(errors.Join(err, errors.New("unable to parse time as RFC3339")))
		return nil
	}
	return &t
}

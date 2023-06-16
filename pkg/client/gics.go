package client

import (
	"bytes"
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/model"
	"fmt"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type GicsClient struct {
	IdentifierSystem string
	RequestUrl       string
}

func NewGicsClient(config config.AppConfig) *GicsClient {
	return &GicsClient{
		RequestUrl:       config.Gics.Fhir.Base + "/$currentConsentForPersonAndTemplate",
		IdentifierSystem: "https://ths-greifswald.de/fhir/gics/identifiers/" + config.Gics.SignerId,
	}
}

func (c *GicsClient) GetConsentStatus(sid model.SignerId, t model.ConsentTemplateKey) (*fhir.Bundle, error) {
	template := fmt.Sprintf("%s;%s;%s", t.DomainName, t.Name, t.Version)
	//default
	ignoreVersionNumber := false

	fhirRequest := fhir.Parameters{
		Id:   nil,
		Meta: nil,
		Parameter: []fhir.ParametersParameter{
			{
				Name:            "personIdentifier",
				ValueIdentifier: &fhir.Identifier{System: &sid.IdType, Value: &sid.Id},
			},
			{
				Name:        "domain",
				ValueString: t.DomainName,
			},
			{
				Name:        "template",
				ValueString: &template,
			},
			{
				Name:         "ignore-version-number",
				ValueBoolean: &ignoreVersionNumber,
			},
		},
	}
	r, err := fhirRequest.MarshalJSON()
	if err != nil {
		return nil, err
	}

	response, err := http.Post(
		c.RequestUrl,
		"application/fhir+json",
		bytes.NewBuffer(r))

	if err != nil {
		log.WithError(err).Error("POST request to gICS failed for: " + c.RequestUrl)
		return nil, err
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.WithError(err).Fatal("Unable to parse gICS get consent status response")
	}
	fmt.Println(string(responseData))
	bundle, err := fhir.UnmarshalBundle(responseData)
	if err != nil {
		log.WithError(err).Fatal("Failed to deserialize FHIR response from  gICS. Expected 'Bundle'")
		return nil, err
	}

	return &bundle, nil
}

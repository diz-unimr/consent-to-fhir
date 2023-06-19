package client

import (
	"bytes"
	"consent-to-fhir/pkg/config"
	"consent-to-fhir/pkg/model"
	"errors"
	"fmt"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type GicsClient struct {
	Auth             *config.Auth
	IdentifierSystem string
	RequestUrl       string
}

func NewGicsClient(config config.AppConfig) *GicsClient {
	client := &GicsClient{
		RequestUrl:       config.Gics.Fhir.Base + "/$currentConsentForPersonAndTemplate",
		IdentifierSystem: "https://ths-greifswald.de/fhir/gics/identifiers/" + config.Gics.SignerId,
	}
	if config.Gics.Fhir.Auth.User != "" && config.Gics.Fhir.Auth.Password != "" {
		client.Auth = &config.Gics.Fhir.Auth
	}

	return client
}

func (c *GicsClient) GetConsentStatus(sid model.SignerId, t model.ConsentTemplateKey) (*fhir.Bundle, error) {
	template := fmt.Sprintf("%s;%s;%s", *t.DomainName, *t.Name, *t.Version)
	//default
	ignoreVersionNumber := false

	fhirRequest := fhir.Parameters{
		Id:   nil,
		Meta: nil,
		Parameter: []fhir.ParametersParameter{
			{
				Name:            "personIdentifier",
				ValueIdentifier: &fhir.Identifier{System: &c.IdentifierSystem, Value: &sid.Id},
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

	response, err := c.postRequest(r)

	if err != nil {
		log.WithError(err).Error("POST request to gICS failed for: " + c.RequestUrl)
		return nil, err
	}
	defer closeBody(response.Body)

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.WithError(err).Fatal("Unable to parse gICS get consent status response")
	}
	if response.StatusCode != http.StatusOK {
		err = errors.New("POST request to gICS failed: " + string(responseData))
		log.WithField("statusCode", response.StatusCode).Error(err.Error())
		return nil, err
	}

	bundle, err := fhir.UnmarshalBundle(responseData)
	if err != nil {
		log.WithError(err).Fatal("Failed to deserialize FHIR response from  gICS. Expected 'Bundle'")
		return nil, err
	}

	return &bundle, nil
}

func (c *GicsClient) postRequest(body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, c.RequestUrl,
		bytes.NewBuffer(body))
	if err != nil {
		log.WithError(err).Fatal("Failed to create POST request")
		return nil, err
	}
	req.Header.Set("Content-Type", "application/fhir+json")
	if c.Auth != nil {
		req.SetBasicAuth(c.Auth.User, c.Auth.Password)
	}

	return http.DefaultClient.Do(req)
}

func closeBody(body io.ReadCloser) {
	err := body.Close()
	if err != nil {
		log.WithError(err).Error("Failed to close response body")
	}
}

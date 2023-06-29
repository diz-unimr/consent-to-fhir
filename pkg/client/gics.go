package client

import (
	"bytes"
	"consent-to-fhir/pkg/config"
	"errors"
	"github.com/samply/golang-fhir-models/fhir-models/fhir"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

type GicsClient struct {
	Auth             *config.Auth
	IdentifierSystem string
	RequestUrl       string
	TargetProfile    string
}

func NewGicsClient(config config.AppConfig) *GicsClient {
	client := &GicsClient{
		RequestUrl:       config.Gics.Fhir.Base + "/$currentPolicyStatesForPerson",
		IdentifierSystem: "https://ths-greifswald.de/fhir/gics/identifiers/" + config.Gics.SignerId,
		TargetProfile:    "https://www.medizininformatik-initiative.de/fhir/modul-consent/StructureDefinition/mii-pr-consent-einwilligung",
	}
	if config.Gics.Fhir.Auth != nil {
		client.Auth = config.Gics.Fhir.Auth
	}

	return client
}

func (c *GicsClient) GetConsentStatus(signerId string, domain string, date string) (*fhir.Bundle, error) {
	//template := fmt.Sprintf("%s;%s;%s", *t.DomainName, *t.Name, *t.Version)
	date = strings.Fields(date)[0]

	//default
	ignoreVersionNumber := false

	fhirRequest := fhir.Parameters{
		Id:   nil,
		Meta: nil,
		Parameter: []fhir.ParametersParameter{
			{
				Name:            "personIdentifier",
				ValueIdentifier: &fhir.Identifier{System: &c.IdentifierSystem, Value: &signerId},
			},
			{
				Name:        "domain",
				ValueString: &domain,
			},
			{
				Name:         "ignore-version-number",
				ValueBoolean: &ignoreVersionNumber,
			},
			{
				Name:      "request-date",
				ValueDate: &date,
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

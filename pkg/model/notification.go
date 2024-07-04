package model

type PolicyState struct {
	Key   *PolicyStateKey `bson:"key" json:"key"`
	Value bool            `bson:"value" json:"value"`
}

type PolicyStateKey struct {
	DomainName *string `bson:"domainName" json:"domainName"`
	Name       *string `bson:"name" json:"name"`
	Version    *string `bson:"version" json:"version"`
}

type Notification struct {
	Context              *Context      `bson:"context" json:"context"`
	ConsentKey           *ConsentKey   `bson:"consentKey" json:"consentKey"`
	PreviousPolicyStates []PolicyState `bson:"previousPolicyStates" json:"previousPolicyStates"`
	CurrentPolicyStates  []PolicyState `bson:"currentPolicyStates" json:"currentPolicyStates"`
}

type Context struct {
	Qc struct {
		QcPassed  bool   `bson:"qcPassed" json:"qcPassed"`
		Type      string `bson:"type json:type"`
		Inspector string `bson:"inspector json:inspector"`
		Comment   string `bson:"comment" json:"comment"`
	} `bson:"qc" json:"qc"`
}

type ConsentKey struct {
	ConsentTemplateKey *ConsentTemplateKey `bson:"consentTemplateKey" json:"consentTemplateKey"`
	SignerIds          []SignerId          `bson:"signerIds" json:"signerIds"`
	ConsentDate        *string             `bson:"consentDate" json:"consentDate"`
}

type ConsentTemplateKey struct {
	DomainName *string `bson:"domainName" json:"domainName"`
	Name       *string `bson:"name" json:"name"`
	Version    *string `bson:"version" json:"version"`
}

type SignerId struct {
	IdType      string `bson:"idType" json:"idType"`
	Id          string `bson:"id" json:"id"`
	OrderNumber *int   `bson:"orderNumber" json:"orderNumber"`
}

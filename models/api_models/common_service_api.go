package api_models

type CopyrightAPI struct {
	LicenseKey        string `json:"licenseKey"`
	PrimaryDomainName string `json:"primaryDomainName"`
	DomainChineseName string `json:"domainChineseName"`
	Project           string `json:"project"`
	Company           string `json:"company"`
	AuditDate         string `json:"auditDate"`
}

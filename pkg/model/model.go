package model

type Cluster struct {
	Name                     string `json:"name"`
	Server                   string `json:"server"`
	InsecureSkipTLSVerify    bool   `json:"insecure-skip-tls-verify"`
	CertificateAuthorityData []byte `json:"certificate-authority-data,omitempty"`
	ClientCertificateData    []byte `json:"client-certificate-data,omitempty"`
	ClientKeyData            []byte `json:"client-key-data,omitempty"`
	BearerToken              string `json:"bearerToken"`
}

type Config struct {
	Shortcuts map[string]string `json:"shortcuts"`
	Clusters  []Cluster         `json:"clusters"`
}

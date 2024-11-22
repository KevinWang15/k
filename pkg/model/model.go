package model

import (
	"encoding/json"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ClusterJSON is an intermediate struct for JSON unmarshaling
type ClusterJSON struct {
	Name                     string          `json:"name"`
	Server                   string          `json:"server"`
	InsecureSkipTLSVerify    bool            `json:"insecure-skip-tls-verify"`
	CertificateAuthorityData []byte          `json:"certificate-authority-data,omitempty"`
	ClientCertificateData    []byte          `json:"client-certificate-data,omitempty"`
	ClientKeyData            []byte          `json:"client-key-data,omitempty"`
	BearerToken              string          `json:"bearerToken"`
	ClusterData              json.RawMessage `json:"cluster,omitempty"`
	UserData                 json.RawMessage `json:"user,omitempty"`
}

// Cluster represents a kubernetes cluster configuration
type Cluster struct {
	Name    string       `json:"name"`
	Cluster *K8sCluster  `json:"cluster,omitempty"`
	User    *K8sAuthInfo `json:"user,omitempty"`
}

// K8sCluster wraps the kubernetes Cluster type to handle the runtime.Object field
type K8sCluster struct {
	Server                   string                     `json:"server"`
	TLSServerName            string                     `json:"tls-server-name,omitempty"`
	InsecureSkipTLSVerify    bool                       `json:"insecure-skip-tls-verify"`
	CertificateAuthority     string                     `json:"certificate-authority,omitempty"`
	CertificateAuthorityData []byte                     `json:"certificate-authority-data,omitempty"`
	ProxyURL                 string                     `json:"proxy-url,omitempty"`
	Extensions               map[string]json.RawMessage `json:"extensions,omitempty"`
}

// K8sAuthInfo wraps the kubernetes AuthInfo type to handle the runtime.Object field
type K8sAuthInfo struct {
	ClientCertificate     string                     `json:"client-certificate,omitempty"`
	ClientCertificateData []byte                     `json:"client-certificate-data,omitempty"`
	ClientKey             string                     `json:"client-key,omitempty"`
	ClientKeyData         []byte                     `json:"client-key-data,omitempty"`
	Token                 string                     `json:"token,omitempty"`
	TokenFile             string                     `json:"tokenFile,omitempty"`
	Impersonate           string                     `json:"act-as,omitempty"`
	ImpersonateGroups     []string                   `json:"act-as-groups,omitempty"`
	ImpersonateUserExtra  map[string][]string        `json:"act-as-user-extra,omitempty"`
	Username              string                     `json:"username,omitempty"`
	Password              string                     `json:"password,omitempty"`
	AuthProvider          *json.RawMessage           `json:"auth-provider,omitempty"`
	Exec                  *api.ExecConfig            `json:"exec,omitempty"`
	Extensions            map[string]json.RawMessage `json:"extensions,omitempty"`
}

type Config struct {
	Shortcuts map[string]string `json:"shortcuts"`
	Clusters  []Cluster         `json:"clusters"`
}

// UnmarshalJSON implements custom JSON unmarshaling for Cluster
func (c *Cluster) UnmarshalJSON(data []byte) error {
	var temp ClusterJSON
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy the simple fields
	c.Name = temp.Name

	// Handle the Cluster field
	if temp.ClusterData != nil {
		c.Cluster = &K8sCluster{}
		if err := json.Unmarshal(temp.ClusterData, c.Cluster); err != nil {
			return err
		}
	}

	// Handle the User field
	if temp.UserData != nil {
		c.User = &K8sAuthInfo{}
		if err := json.Unmarshal(temp.UserData, c.User); err != nil {
			return err
		}
	}

	return nil
}

// ToAPICluster converts K8sCluster to api.Cluster
func (k *K8sCluster) ToAPICluster() *api.Cluster {
	if k == nil {
		return nil
	}
	return &api.Cluster{
		Server:                   k.Server,
		TLSServerName:            k.TLSServerName,
		InsecureSkipTLSVerify:    k.InsecureSkipTLSVerify,
		CertificateAuthority:     k.CertificateAuthority,
		CertificateAuthorityData: k.CertificateAuthorityData,
		ProxyURL:                 k.ProxyURL,
		Extensions:               make(map[string]runtime.Object), // Initialize empty extensions
	}
}

// FromAPICluster converts api.Cluster to K8sCluster
func FromAPICluster(c *api.Cluster) *K8sCluster {
	if c == nil {
		return nil
	}

	extensions := make(map[string]json.RawMessage)
	for k, v := range c.Extensions {
		if data, err := json.Marshal(v); err == nil {
			extensions[k] = data
		}
	}

	return &K8sCluster{
		Server:                   c.Server,
		TLSServerName:            c.TLSServerName,
		InsecureSkipTLSVerify:    c.InsecureSkipTLSVerify,
		CertificateAuthority:     c.CertificateAuthority,
		CertificateAuthorityData: c.CertificateAuthorityData,
		ProxyURL:                 c.ProxyURL,
		Extensions:               extensions,
	}
}

// ToAPIAuthInfo converts K8sAuthInfo to api.AuthInfo
func (k *K8sAuthInfo) ToAPIAuthInfo() *api.AuthInfo {
	if k == nil {
		return nil
	}
	return &api.AuthInfo{
		ClientCertificate:     k.ClientCertificate,
		ClientCertificateData: k.ClientCertificateData,
		ClientKey:             k.ClientKey,
		ClientKeyData:         k.ClientKeyData,
		Token:                 k.Token,
		TokenFile:             k.TokenFile,
		Impersonate:           k.Impersonate,
		ImpersonateGroups:     k.ImpersonateGroups,
		ImpersonateUserExtra:  k.ImpersonateUserExtra,
		Username:              k.Username,
		Password:              k.Password,
		Exec:                  k.Exec,
		Extensions:            make(map[string]runtime.Object), // Initialize empty extensions
	}
}

// FromAPIAuthInfo converts api.AuthInfo to K8sAuthInfo
func FromAPIAuthInfo(a *api.AuthInfo) *K8sAuthInfo {
	if a == nil {
		return nil
	}

	extensions := make(map[string]json.RawMessage)
	for k, v := range a.Extensions {
		if data, err := json.Marshal(v); err == nil {
			extensions[k] = data
		}
	}

	var authProvider *json.RawMessage
	if a.AuthProvider != nil {
		if data, err := json.Marshal(a.AuthProvider); err == nil {
			authProvider = (*json.RawMessage)(&data)
		}
	}

	return &K8sAuthInfo{
		ClientCertificate:     a.ClientCertificate,
		ClientCertificateData: a.ClientCertificateData,
		ClientKey:             a.ClientKey,
		ClientKeyData:         a.ClientKeyData,
		Token:                 a.Token,
		TokenFile:             a.TokenFile,
		Impersonate:           a.Impersonate,
		ImpersonateGroups:     a.ImpersonateGroups,
		ImpersonateUserExtra:  a.ImpersonateUserExtra,
		Username:              a.Username,
		Password:              a.Password,
		AuthProvider:          authProvider,
		Exec:                  a.Exec,
		Extensions:            extensions,
	}
}

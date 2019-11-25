package config

import "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"

type ClusterConfig struct {
	OperatorImage string
	Labels        map[string]string
	Name          string
	PXC           Spec
	ProxySQL      Spec
	S3            k8s.S3StorageConfig
	PMM           PMMSpec
}

type Spec struct {
	StorageSize     string
	StorageClass    string
	Instances       int32
	RequestCPU      string
	RequestMem      string
	AntiAffinityKey string
	BrokerInstance  string
}

type PMMSpec struct {
	Enabled    bool   `json:"enabled,omitempty"`
	ServerHost string `json:"serverHost,omitempty"`
	Image      string `json:"image,omitempty"`
	ServerUser string `json:"serverUser,omitempty"`
	ServerPass string `json:"serverPass,omitempty"`
}

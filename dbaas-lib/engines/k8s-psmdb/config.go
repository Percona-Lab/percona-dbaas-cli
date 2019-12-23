package psmdb

import (
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/config"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
)

// PSMDBCluster represent interface for ckuster types
type PSMDBCluster interface {
	UpdateWith(c config.ClusterConfig, s3 *k8s.BackupStorageSpec) (err error)
	Upgrade(imgs map[string]string)
	SetNew(c config.ClusterConfig, s3 *k8s.BackupStorageSpec, p k8s.PlatformType) (err error)
	GetName() string
	MarshalRequests() error
	GetCR() (string, error)
	SetLabels(labels map[string]string)
}

type AppState string

const (
	AppStateUnknown AppState = "unknown"
	AppStateInit             = "initializing"
	AppStateReady            = "ready"
	AppStateError            = "error"
)

type PSMDBClusterStatus struct {
	Messages []string `json:"message,omitempty"`
	Status   AppState `json:"state,omitempty"`
}

type AppStatus struct {
	Size    int32    `json:"size,omitempty"`
	Ready   int32    `json:"ready,omitempty"`
	Status  AppState `json:"status,omitempty"`
	Message string   `json:"message,omitempty"`
}

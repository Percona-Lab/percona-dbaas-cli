package psmdb

import "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"

type PSMDBCluster interface {
	Upgrade(imgs map[string]string)
	GetName() string
	SetName(name string)
	SetUsersSecretName(name string)
	MarshalRequests() error
	GetCR() (string, error)
	SetLabels(labels map[string]string)
	GetOperatorImage() string
	SetDefaults() error
	SetupMiniConfig()
	GetStatus() dbaas.State
	GetReplestsNames() []string
}

package pdl

import (
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
)

type Engine interface {
	ParseOptions(s string) error
	CreateDBCluster(name string) error
	CheckDBClusterStatus(name string) (structs.DB, error)
	DeleteDBCluster(name string, delePVC bool) error
	GetDBCluster(name string) (structs.DB, error)
	UpdateDBCluster() error
	ListDBClusters() error
	DescribeDBCluster(name string) error
}

var Providers = make(map[string]Provider)

type Provider struct {
	Engines map[string]Engine
}

func RegisterEngine(providerName, engineName string, eng Engine) {
	if _, ok := Providers[providerName].Engines[engineName]; !ok && Providers[providerName].Engines == nil {
		engns := map[string]Engine{
			engineName: eng,
		}
		Providers[providerName] = Provider{
			Engines: engns,
		}
	}
	Providers[providerName].Engines[engineName] = eng
}

package pdl

import (
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
)

type Engine interface {
	ParseOptions(opts string) error
	CreateDBCluster(name, opts string) error
	DeleteDBCluster(name, opts string, delePVC bool) (string, error)
	GetDBCluster(name, opts string) (structs.DB, error)
	GetDBClusterList() ([]structs.DB, error)
	UpdateDBCluster(name, opts string) error
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

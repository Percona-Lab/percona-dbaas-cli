package dbaas

import (
	"errors"

	_ "github.com/Percona-Lab/percona-dbaas-cli/dbaas/engines/pxc"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pdl"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/structs"
)

type Instance struct {
	Name          string
	Engine        string
	Provider      string
	ClusterSize   int
	DiskSize      string
	EngineOptions string
}

func CreateDB(instance Instance) (*structs.DB, error) {
	if _, providerOk := pdl.Providers[instance.Provider]; !providerOk {
		return nil, errors.New("wrong engine")
	}
	if _, ok := pdl.Providers[instance.Provider].Engines[instance.Engine]; !ok {
		return nil, errors.New("wrong engine")
	}

	err := pdl.Providers[instance.Provider].Engines[instance.Engine].ParseOptions(instance.EngineOptions)
	if err != nil {
		return nil, err
	}
	err = pdl.Providers[instance.Provider].Engines[instance.Engine].CreateDBCluster(instance.Name)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func DescribeDB(instance Instance) ([]structs.DB, error) {
	return nil, nil
}

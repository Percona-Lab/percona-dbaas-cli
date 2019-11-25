package dbaas

import (
	"errors"

	_ "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/pxc"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/pdl"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
)

type Instance struct {
	Name          string
	Engine        string
	Provider      string
	ClusterSize   int
	DiskSize      string
	EngineOptions string
}

func (i *Instance) setDefaults() {
	if len(i.Engine) == 0 {
		i.Engine = "pxc"
	}
	if len(i.Provider) == 0 {
		i.Provider = "k8s"
	}
}

func CreateDB(instance Instance) (*structs.DB, error) {
	instance.setDefaults()
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

func DeleteDB(instance Instance) error {
	instance.setDefaults()
	err := pdl.Providers[instance.Provider].Engines[instance.Engine].ParseOptions(instance.EngineOptions)
	if err != nil {
		return err
	}
	err = pdl.Providers[instance.Provider].Engines[instance.Engine].DeleteDBCluster(instance.Name, false)
	if err != nil {
		return err
	}
	return nil
}

func DescribeDB(instance Instance) ([]structs.DB, error) {
	return nil, nil
}

package dbaas

import (
	"errors"

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

// CreateDB creates DB resource using name, provider, engine and options given in 'instance' object. The default value provider=k8s, engine=pxc
func CreateDB(instance Instance) error {
	err := checkPrroviderAndEngine(instance)
	if err != nil {
		return err
	}
	err = pdl.Providers[instance.Provider].Engines[instance.Engine].CreateDBCluster(instance.Name, instance.EngineOptions)
	if err != nil {
		return err
	}

	return nil
}

// ModifyDB modifies DB resource using name, provider, engine and options given in 'instance' object. The default value provider=k8s, engine=pxc
func ModifyDB(instance Instance) error {
	err := checkPrroviderAndEngine(instance)
	if err != nil {
		return err
	}
	err = pdl.Providers[instance.Provider].Engines[instance.Engine].UpdateDBCluster(instance.Name, instance.EngineOptions)
	if err != nil {
		return err
	}

	return nil
}

func DescribeDB(instance Instance) (structs.DB, error) {
	err := checkPrroviderAndEngine(instance)
	if err != nil {
		return structs.DB{}, err
	}

	return pdl.Providers[instance.Provider].Engines[instance.Engine].GetDBCluster(instance.Name, instance.EngineOptions)
}

func ListDB(instance Instance) ([]structs.DB, error) {
	err := checkPrroviderAndEngine(instance)
	if err != nil {
		return nil, err
	}

	return pdl.Providers[instance.Provider].Engines[instance.Engine].GetDBClusterList()
}

func DeleteDB(instance Instance, saveData bool) (string, error) {
	err := checkPrroviderAndEngine(instance)
	if err != nil {
		return "", err
	}

	return pdl.Providers[instance.Provider].Engines[instance.Engine].DeleteDBCluster(instance.Name, instance.EngineOptions, saveData)
}

func checkPrroviderAndEngine(instance Instance) error {
	if _, providerOk := pdl.Providers[instance.Provider]; !providerOk {
		return errors.New("wrong provider")
	}
	if _, ok := pdl.Providers[instance.Provider].Engines[instance.Engine]; !ok {
		return errors.New("wrong engine")
	}
	return nil
}

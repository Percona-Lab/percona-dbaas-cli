package dbaas

import (
	"github.com/pkg/errors"
)

type Instance struct {
	Name          string
	Engine        string
	Provider      string
	ClusterSize   int
	DiskSize      string
	EngineOptions string
	RootPass      string
	Version       string
}

// CreateDB creates DB resource using name, provider, engine and options given in 'instance' object. The default value provider=k8s, engine=pxc
func CreateDB(instance Instance) error {
	err := checkProviderAndEngine(instance)
	if err != nil {
		return err
	}

	err = Providers[instance.Provider].Engines[instance.Engine].CreateDBCluster(instance.Name, instance.EngineOptions, instance.RootPass, instance.Version)
	if err != nil {
		return err
	}

	return nil
}

// ModifyDB modifies DB resource using name, provider, engine and options given in 'instance' object. The default value provider=k8s, engine=pxc
func ModifyDB(instance Instance) error {
	err := checkProviderAndEngine(instance)
	if err != nil {
		return err
	}

	err = Providers[instance.Provider].Engines[instance.Engine].UpdateDBCluster(instance.Name, instance.EngineOptions, instance.Version)
	if err != nil {
		return err
	}

	return nil
}

func DescribeDB(instance Instance) (DB, error) {
	err := checkProviderAndEngine(instance)
	if err != nil {
		return DB{}, err
	}

	return Providers[instance.Provider].Engines[instance.Engine].GetDBCluster(instance.Name, instance.EngineOptions)
}

func ListDB(instance Instance) ([]DB, error) {
	err := checkProviderAndEngine(instance)
	if err != nil {
		return nil, err
	}

	return Providers[instance.Provider].Engines[instance.Engine].GetDBClusterList()
}

func DeleteDB(instance Instance, saveData bool) (string, error) {
	err := checkProviderAndEngine(instance)
	if err != nil {
		return "", err
	}

	return Providers[instance.Provider].Engines[instance.Engine].DeleteDBCluster(instance.Name, instance.EngineOptions, instance.Version, saveData)
}

func checkProviderAndEngine(instance Instance) error {
	if _, providerOk := Providers[instance.Provider]; !providerOk {
		return errors.New("wrong provider")
	}
	if _, ok := Providers[instance.Provider].Engines[instance.Engine]; !ok {
		return errors.New("wrong engine")
	}

	return nil
}

func PreCheck(instance Instance) ([]string, error) {
	err := checkProviderAndEngine(instance)
	if err != nil {
		return nil, err
	}

	return Providers[instance.Provider].Engines[instance.Engine].PreCheck(instance.Name, instance.EngineOptions, instance.Version)
}

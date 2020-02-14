package dbaas

import (
	"github.com/pkg/errors"

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
	RootPass      string
}

// CreateDB creates DB resource using name, provider, engine and options given in 'instance' object. The default value provider=k8s, engine=pxc
func CreateDB(instance Instance) error {
	err := checkPrroviderAndEngine(instance)
	if err != nil {
		return err
	}
	err = pdl.Providers[instance.Provider].Engines[instance.Engine].CreateDBCluster(instance.Name, instance.EngineOptions, instance.RootPass)
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

/*
const (
	PlatformKubernetes PlatformType = "kubernetes"
	PlatformMinikube   PlatformType = "minikube"
	PlatformOpenshift  PlatformType = "openshift"
	PlatformMinishift  PlatformType = "minishift"
)

type PlatformType string

// GetPlatformType is for determine and return platform type
func (i Instance) GetPlatformType() {
	if checkMinikube() {
		i.Platform = PlatformMinikube
		return
	}

	if checkMinishift() {
		i.Platform = PlatformMinishift
		return
	}

	if checkOpenshift() {
		i.Platform = PlatformOpenshift
		return
	}

	i.Platform = PlatformKubernetes
}

func checkMinikube() bool {
	output, err := runCmd("kubectl", "get", "storageclass", "-o", "jsonpath='{.items..provisioner}'")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "k8s.io/minikube-hostpath")
}

func checkMinishift() bool {
	output, err := runCmd("kubectl", "get", "pods", "master-etcd-localhost", "-n", "kube-system", "-o", "jsonpath='{.spec.volumes..path}'")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "minishift")
}

func checkOpenshift() bool {
	output, err := runCmd("kubectl", "api-versions")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "openshift")
}
func runCmd(cmd string, args ...string) ([]byte, error) {
	cli := exec.Command(cmd, args...)
	cli.Env = os.Environ()

	o, err := cli.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "run command")
	}

	return o, nil
}
*/

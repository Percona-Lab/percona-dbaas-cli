package psmdb

import (
	"fmt"
	"os"
	"strings"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/config"

	v110 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/v110"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/pdl"
	"github.com/pkg/errors"
)

const (
	defaultOperatorVersion = "percona/percona-server-mongodb-operator:1.1.0"
	provider               = "k8s"
	engine                 = "psmdb"
)

var objects map[Version]VersionObject

func init() {
	// Register psmdb engine in dbaas
	psmdb, err := NewPSMDBController("", "k8s")
	if err != nil {
		fmt.Println("Cant start. Setup your kubectl")
		os.Exit(1)
	}

	pdl.RegisterEngine(provider, engine, psmdb)

	// Register psmdb versions
	objects = make(map[Version]VersionObject)
	objects[currentVersion] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v110.Bundle,
		},
		psmdb: &v110.PerconaServerMongoDB{},
	}
}

// PSMDB represents PSMDB Operator controller
type PSMDB struct {
	cmd    *k8s.Cmd
	config config.ClusterConfig
}

const (
	currentVersion Version = "default"
)

type Version string

type PSMDBMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type PSMDBResource struct {
	Meta   PSMDBMeta `json:"metadata"`
	Status PSMDBClusterStatus
}
type k8sCluster struct {
	Items []PSMDBResource `json:"items"`
}

type k8sStatus struct {
	Status PSMDBClusterStatus
}

type PVCMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	SelfLink  string `json:"selflink"`
	UID       string `json:"uid"`
}

type k8sPVC struct {
	Meta PVCMeta `json:"metadata"`
}

type VersionObject struct {
	k8s   k8s.Objects
	psmdb PSMDBCluster
}

// NewPSMDBController returns new PSMDBOperator Controller
func NewPSMDBController(envCrt, provider string) (*PSMDB, error) {
	var psmdb PSMDB
	if len(provider) == 0 || provider == "k8s" {
		k8sCmd, err := k8s.New(envCrt)
		if err != nil {
			return nil, errors.Wrap(err, "new Cmd")
		}
		psmdb.cmd = k8sCmd
	}
	return &psmdb, nil
}

func (p PSMDB) bundle(v map[Version]VersionObject, operatorVersion string) []k8s.BundleObject {
	if operatorVersion == "" {
		operatorVersion = defaultOperatorVersion
	}

	for i, o := range v[currentVersion].k8s.Bundle {
		if o.Kind == "Deployment" && o.Name == p.operatorName() {
			v[currentVersion].k8s.Bundle[i].Data = strings.Replace(o.Data, "{{image}}", operatorVersion, -1)
		}
	}
	return v[currentVersion].k8s.Bundle
}

func (p PSMDB) getCR(cluster PSMDBCluster) (string, error) {
	return cluster.GetCR()
}

func (p *PSMDB) setup(cluster PSMDBCluster, c config.ClusterConfig, s3 *k8s.BackupStorageSpec, platform k8s.PlatformType) error {
	err := cluster.SetNew(c, s3, platform)
	if err != nil {
		return errors.Wrap(err, "parse options")
	}

	err = cluster.MarshalRequests()
	if err != nil {
		return errors.Wrap(err, "marshal psmdb volume requests")
	}

	return nil
}

func (p *PSMDB) operatorName() string {
	return "percona-server-mongodb-operator"
}

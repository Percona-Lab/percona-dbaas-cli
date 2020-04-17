package psmdb

import (
	"fmt"
	"os"

	v110 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/v110"
	v120 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/v120"
	v130 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/v130"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/pdl"
	"github.com/pkg/errors"
)

const (
	provider       = "k8s"
	engine         = "psmdb"
	defaultVersion = "1.3.0"
)

type Version string

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
	objects["1.1.0"] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v110.Bundle,
		},
		psmdb: &v110.PerconaServerMongoDB{},
	}
	objects["1.2.0"] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v120.Bundle,
		},
		psmdb: &v120.PerconaServerMongoDB{},
	}
	objects["1.3.0"] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v130.Bundle,
		},
		psmdb: &v130.PerconaServerMongoDB{},
	}
}

// PSMDB represents PSMDB Operator controller
type PSMDB struct {
	cmd          *k8s.Cmd
	conf         PSMDBCluster
	platformType k8s.PlatformType
	bundle       []k8s.BundleObject
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
		psmdb.platformType = k8sCmd.GetPlatformType()
	}
	return &psmdb, nil
}

func (p *PSMDB) setVersionObjectsWithDefaults(version Version) error {
	if p.conf != nil && p.bundle != nil {
		return nil
	}
	if len(version) == 0 {
		version = defaultVersion
	}
	if _, ok := objects[version]; !ok {
		return errors.Errorf("unsupporeted version %s", version)
	}

	p.conf = objects[version].psmdb
	err := p.conf.SetDefaults()
	if err != nil {
		errors.Wrap(err, "set defaults")
	}
	p.bundle = objects[version].k8s.Bundle

	return nil
}

func (p PSMDB) getCR(cluster PSMDBCluster) (string, error) {
	return cluster.GetCR()
}

func (p *PSMDB) operatorName() string {
	return "percona-server-mongodb-operator"
}

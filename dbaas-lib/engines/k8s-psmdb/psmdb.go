package psmdb

import (
	"fmt"
	"os"
	"strings"

	v110 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/v110"
	v120 "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb/types/v120"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/pdl"
	"github.com/pkg/errors"
)

const (
	provider               = "k8s"
	engine                 = "psmdb"
	defaultVersion Version = "1.1.0"
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
}

// PSMDB represents PSMDB Operator controller
type PSMDB struct {
	cmd  *k8s.Cmd
	conf PSMDBCluster
}

type Version string

type PSMDBMeta struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type AppState string

const (
	AppStateUnknown AppState = "unknown"
	AppStateInit             = "initializing"
	AppStateReady            = "ready"
	AppStateError            = "error"
)

type PSMDBClusterStatus struct {
	Messages []string `json:"message,omitempty"`
	Status   AppState `json:"state,omitempty"`
}

type AppStatus struct {
	Size    int32    `json:"size,omitempty"`
	Ready   int32    `json:"ready,omitempty"`
	Status  AppState `json:"status,omitempty"`
	Message string   `json:"message,omitempty"`
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
		operatorVersion = v[defaultVersion].psmdb.GetOperatorImage()
	}

	for i, o := range v[defaultVersion].k8s.Bundle {
		if o.Kind == "Deployment" && o.Name == p.operatorName() {
			v[defaultVersion].k8s.Bundle[i].Data = strings.Replace(o.Data, "{{image}}", operatorVersion, -1)
		}
	}
	return v[defaultVersion].k8s.Bundle
}

func (p PSMDB) getCR(cluster PSMDBCluster) (string, error) {
	return cluster.GetCR()
}

func (p *PSMDB) operatorName() string {
	return "percona-server-mongodb-operator"
}

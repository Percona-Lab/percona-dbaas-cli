package pxc

import (
	"encoding/json"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/engines/pxc/types/config"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pdl"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/structs"
	"github.com/pkg/errors"
)

const (
	defaultVersion = "default"
)

// PXC represents PXC Operator controller
type PXC struct {
	cmd    *k8s.Cmd
	config config.ClusterConfig
}

func init() {
	pxc, err := NewPXCController("", "k8s")
	if err != nil {
		return
	}
	pdl.RegisterEngine("k8s", "pxc", pxc)
}

// NewPXCController returns new PXCOperator Controller
func NewPXCController(envCrt, provider string) (*PXC, error) {
	var pxc PXC
	if len(provider) == 0 || provider == "k8s" {
		k8sCmd, err := k8s.New(envCrt)
		if err != nil {
			return nil, errors.Wrap(err, "new Cmd")
		}
		pxc.cmd = k8sCmd
	}
	return &pxc, nil
}

func (p *PXC) ParseOptions(s string) error {
	return nil
}

// CreateDBCluster start creating DB cluster
func (p *PXC) CreateDBCluster(name string) error {
	var s3stor *k8s.BackupStorageSpec
	c := objects[currentVersion].pxc
	err := p.setup(c, p.config, s3stor, p.cmd.GetPlatformType())
	if err != nil {
		return errors.Wrap(err, "set configuration: ")
	}
	cr, err := p.getCR(c)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.CreateCluster("pxc", p.config.OperatorImage, name, cr, p.bundle(objects, p.config.OperatorImage))
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

// CheckDBClusterStatus status return Cluster object with cluster info
func (p *PXC) CheckDBClusterStatus(name string) (structs.DB, error) {
	var db structs.DB
	secrets, err := p.cmd.GetSecrets(name)
	if err != nil {
		return db, errors.Wrap(err, "get cluster secrets")

	}
	status, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return db, errors.Wrap(err, "get cluster status")

	}
	st := &k8sStatus{}
	err = json.Unmarshal(status, st)
	if err != nil {
		return db, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.Status {
	case AppStateReady:
		db.Host = st.Status.Host
		db.Port = 3306
		db.User = "root"
		db.Pass = string(secrets["root"])
		db.Status = k8s.ClusterStateReady
		return db, nil
	case AppStateInit:
		db.Status = k8s.ClusterStateInit
		return db, nil
	case AppStateError:
		db.Status = k8s.ClusterStateError
		return db, errors.New(st.Status.Messages[0])
	default:
		return db, errors.New("unknown status")
	}
}

// DeleteDBCluster delete cluster by name
func (p *PXC) DeleteDBCluster(name string, delePVC bool) error {
	ext, err := p.cmd.IsObjExists("pxc", name)
	if err != nil {
		return errors.Wrap(err, "check if cluster exists")
	}

	if !ext {
		return errors.New("unable to find cluster pxc/" + name)
	}

	err = p.cmd.DeleteCluster("pxc", p.operatorName(), name, delePVC)
	if err != nil {
		return errors.Wrap(err, "delete cluster")
	}
	return nil
}

func (p *PXC) GetDBCluster(name string) (structs.DB, error) {
	return structs.DB{}, nil
}

func (p *PXC) UpdateDBCluster() error {
	return nil
}

func (p *PXC) ListDBClusters() error {
	return nil
}

func (p *PXC) DescribeDBCluster(name string) error {
	return nil
}

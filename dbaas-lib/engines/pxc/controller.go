package pxc

import (
	"encoding/json"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/pxc/types/config"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/structs"
	"github.com/pkg/errors"
)

func (p *PXC) ParseOptions(s string) error {
	var c config.ClusterConfig
	if len(s) == 0 {
		c.PXC.Instances = int32(3)
		c.PXC.RequestCPU = "600m"
		c.PXC.RequestMem = "1G"
		c.PXC.AntiAffinityKey = "kubernetes.io/hostname"

		c.ProxySQL.Instances = int32(1)
		c.ProxySQL.RequestCPU = "600m"
		c.ProxySQL.RequestMem = "1G"
		c.ProxySQL.AntiAffinityKey = "kubernetes.io/hostname"

		c.S3.SkipStorage = true
		p.config = c
		return nil
	}
	return nil
}

// CreateDBCluster start creating DB cluster
func (p *PXC) CreateDBCluster(name string) error {
	var s3stor *k8s.BackupStorageSpec
	c := objects[currentVersion].pxc
	p.config.Name = name
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

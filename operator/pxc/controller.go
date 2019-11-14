package pxc

import (
	"encoding/json"

	"github.com/Percona-Lab/percona-dbaas-cli/operator/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/operator/pxc/types/config"
	"github.com/pkg/errors"
)

const (
	defaultVersion = "default"
)

// PXC represents PXC Operator controller
type PXC struct {
	cmd *k8s.Cmd
}

// NewPXCController returns new PXCOperator Controller
func NewPXCController(labels map[string]string, envCrt string) (*PXC, error) {
	cmd, err := k8s.New(envCrt)
	if err != nil {
		return nil, errors.Wrap(err, "new Cmd")
	}

	return &PXC{
		cmd: cmd,
	}, nil
}

// CreateDBCluster start creating DB cluster
func (p *PXC) CreateDBCluster(config config.ClusterConfig, operatorImage string) error {
	var s3stor *k8s.BackupStorageSpec
	c := objects[currentVersion].pxc
	err := p.setup(c, config, s3stor, p.cmd.GetPlatformType())
	if err != nil {
		return errors.Wrap(err, "set configuration: ")
	}
	cr, err := p.getCR(c)
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.CreateCluster("pxc", operatorImage, c.GetName(), cr, p.bundle(objects, operatorImage))
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

// CheckDBClusterStatus status return Cluster object with cluster info
func (p *PXC) CheckDBClusterStatus(name string) (Cluster, error) {
	secrets, err := p.cmd.GetSecrets(name)
	if err != nil {
		return Cluster{}, errors.Wrap(err, "get cluster secrets")

	}
	status, err := p.cmd.GetObject("pxc", name)
	if err != nil {
		return Cluster{}, errors.Wrap(err, "get cluster status")

	}
	st := &k8sStatus{}
	cluster := Cluster{}
	err = json.Unmarshal(status, st)
	if err != nil {
		return cluster, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.Status {
	case AppStateReady:
		cluster.Host = st.Status.Host
		cluster.Port = 3306
		cluster.User = "root"
		cluster.Pass = string(secrets["root"])
		cluster.State = k8s.ClusterStateReady
		return cluster, nil
	case AppStateInit:
		cluster.State = k8s.ClusterStateInit
		return cluster, nil
	case AppStateError:
		cluster.State = k8s.ClusterStateError
		return cluster, errors.New(st.Status.Messages[0])
	}
	cluster.State = k8s.ClusterStateUnknown
	return cluster, nil
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

func (p *PXC) GetDBCluster(name string) (Cluster, error) {
	return Cluster{}, nil
}

func (p *PXC) UpdateDBCluster(config config.ClusterConfig) error {
	return nil
}

func (p *PXC) ListDBClusters() error {
	return nil
}

func (p *PXC) DescribeDBCluster(name string) error {
	return nil
}

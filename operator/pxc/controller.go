package pxc

import (
	"encoding/json"
	"fmt"

	"github.com/Percona-Lab/percona-dbaas-cli/operator/k8s"
	"github.com/pkg/errors"
)

const (
	defaultVersion = "default"
)

// PXC represents PXC Operator controller
type PXC struct {
	cmd    *k8s.Cmd
	config *PerconaXtraDBCluster
	obj    k8s.Objects
}

// NewController returns new PXCOperator Controller
func NewController(labels map[string]string, envCrt string) (*PXC, error) {
	config := &PerconaXtraDBCluster{}
	cmd, err := k8s.New(envCrt)
	if err != nil {
		return nil, errors.Wrap(err, "new Cmd")
	}
	config.ObjectMeta.Labels = labels

	return &PXC{
		cmd:    cmd,
		obj:    Objects[defaultVersion],
		config: config,
	}, nil
}

// CreateCluster start creating cluster procces
func (p *PXC) CreateCluster(config ClusterConfig, operatorImage string) error {
	var s3stor *k8s.BackupStorageSpec
	err := p.Setup(config, s3stor, p.cmd.GetPlatformType())
	if err != nil {
		return errors.Wrap(err, "set configuration: ")
	}
	cr, err := p.getCR()
	if err != nil {
		return errors.Wrap(err, "get cr")
	}

	err = p.cmd.CreateCluster("pxc", operatorImage, p.config.ObjectMeta.Name, cr, p.bundle(operatorImage))
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

type Cluster struct {
	Host  string           `json:"host,omitempty"`
	Port  int              `json:"port,omitempty"`
	User  string           `json:"user,omitempty"`
	Pass  string           `json:"pass,omitempty"`
	State k8s.ClusterState `json:"state,omitempty"`
}

func (c Cluster) String() string {
	stringMsg := `Host: %s, Port: 3306, User: root, Pass: %s`
	return fmt.Sprintf(stringMsg, c.Host, c.Pass)
}

// CheckClusterStatus status return Cluster object with cluster info
func (p *PXC) CheckClusterStatus() (Cluster, error) {
	secrets, err := p.cmd.GetSecrets(p.config.ObjectMeta.Name)
	if err != nil {
		return Cluster{}, errors.Wrap(err, "get cluster secrets")

	}
	status, err := p.cmd.GetObject("pxc", p.config.ObjectMeta.Name)
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

// DeleteCluster delete cluster by name
func (p *PXC) DeleteCluster(name string, delePVC bool) error {
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

func (p *PXC) GetCluster(name string) (Cluster, error) {
	return Cluster{}, nil
}

func (p *PXC) UpdateCluster(config ClusterConfig) error {
	return nil
}

func (p *PXC) ListClusters() error {
	return nil
}

func (p *PXC) Describe() error {
	return nil
}

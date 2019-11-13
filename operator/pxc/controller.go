package pxc

import (
	"regexp"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/operator/dbaas"
	"github.com/pkg/errors"
)

const (
	defaultVersion    = "default"
	getStatusMaxTries = 1200
)

// Controller represents PXC Operator controller
type Controller struct {
	cmd *dbaas.Cmd
	app *pxc
}

// NewController returns new PXCOperator Controller
func NewController(labels, envCrt, name string) (Controller, error) {
	cmd, err := dbaas.New(envCrt)
	if err != nil {
		return Controller{}, errors.Wrap(err, "new Cmd")
	}
	if len(labels) > 0 { //pass map to method
		match, err := regexp.MatchString("^(([a-zA-Z0-9_]+=[a-zA-Z0-9_]+)(,|$))+$", labels)
		if err != nil {
			return Controller{}, errors.Wrap(err, "label parse")
		}
		if !match {
			return Controller{}, errors.New("Incorrect label format. Use key1=value1,key2=value2 syntax")

		}
	}

	app := new(name, defaultVersion, labels)

	return Controller{
		cmd: cmd,
		app: app,
	}, nil
}

// CreateCluster start creating cluster procces
func (c *Controller) CreateCluster(config ClusterConfig, skipS3Storage bool, operatorImage string) error {
	var s3stor *dbaas.BackupStorageSpec
	if !skipS3Storage {
		var err error
		s3stor, err = c.cmd.S3Storage(c.app.name, config.S3)
		if err != nil {
			switch err.(type) {
			case dbaas.ErrNoS3Options:
				return errors.Wrap(err, "no S3 storage")
			default:
				return errors.Wrap(err, "create S3 backup storage: ")
			}
		}
	}
	err := c.app.Setup(config, s3stor, c.cmd.GetPlatformType())
	if err != nil {
		return errors.Wrap(err, "set configuration: ")
	}
	cr, err := c.app.getCr()
	if err != nil {
		return errors.Wrap(err, "get cr")
	}
	err = c.cmd.CreateCluster("pxc", operatorImage, c.app.name, cr, c.app.bundle(operatorImage))
	if err != nil {
		return errors.Wrap(err, "create cluster")
	}

	return nil
}

// CheckClusterReady wait for cluster state ready
func (c *Controller) CheckClusterReady() (Cluster, error) {
	// give a time for operator to start
	time.Sleep(1 * time.Minute)

	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		secrets, err := c.cmd.GetSecrets(c.app.name)
		if err != nil {
			return Cluster{}, errors.Wrap(err, "get cluster secrets")

		}
		status, err := c.cmd.GetObject("pxc", c.app.name)
		if err != nil {
			return Cluster{}, errors.Wrap(err, "get cluster status")

		}
		cluster, err := c.app.CheckClusterStatus(status, secrets)
		if err != nil {
			return Cluster{}, errors.Wrap(err, "parse cluster status")

		}

		switch cluster.State {
		case dbaas.ClusterStateReady:
			return cluster, nil

		case dbaas.ClusterStateInit:
		}

		if tries >= getStatusMaxTries {
			return cluster, errors.New("unable to start cluster")

		}
		tries++
	}
	return Cluster{}, errors.New("unable to start cluster")
}

// DeleteCluster delete cluster by name
func (c *Controller) DeleteCluster(name string, delePVC bool) error {
	ext, err := c.cmd.IsObjExists("pxc", name)
	if err != nil {
		return errors.Wrap(err, "check if cluster exists")
	}

	if !ext {
		return errors.New("unable to find cluster pxc/" + name)
	}

	err = c.cmd.DeleteCluster("pxc", c.app.operatorName(), c.app.name, delePVC)
	if err != nil {
		return errors.Wrap(err, "delete cluster")
	}
	return nil
}

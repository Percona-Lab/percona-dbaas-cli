package pxc

import (
	"regexp"

	"github.com/Percona-Lab/percona-dbaas-cli/operator/dbaas"
	"github.com/pkg/errors"
)

const (
	defaultVersion = "default"
)

type Controller struct {
	Cmd *dbaas.Cmd
	App *PXC
}

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

	app := New(name, defaultVersion, labels)

	return Controller{
		Cmd: cmd,
		App: app,
	}, nil
}

func (c *Controller) CreatDB(config ClusterConfig, skipS3Storage bool, operatorImage string) (dbaas.Msg, error) {
	var s3stor *dbaas.BackupStorageSpec
	if !skipS3Storage {
		var err error
		s3stor, err = c.Cmd.S3Storage(c.App, config.S3)
		if err != nil {
			switch err.(type) {
			case dbaas.ErrNoS3Options:
				return &Msg{}, errors.Wrap(err, "no S3 storage")
			default:
				return &Msg{}, errors.Wrap(err, "create S3 backup storage: ")
			}
		}
	}
	_, err := c.App.Setup(config, s3stor, c.Cmd.GetPlatformType())
	if err != nil {
		return &Msg{}, errors.Wrap(err, "set configuration: ")
	}
	created := make(chan dbaas.Msg)
	msg := make(chan dbaas.OutuputMsg)
	cerr := make(chan error)

	go c.Cmd.Create("pxc", c.App, operatorImage, created, msg, cerr)

	for {
		select {
		case okmsg := <-created:
			return okmsg, nil
		case omsg := <-msg:
			switch omsg.(type) {
			case dbaas.OutuputMsgDebug:
				// fmt.Printf("\n[debug] %s\n", omsg)
			case dbaas.OutuputMsgError:
				//return &Msg{}, errors.Wrap(err, "operator log error")
			}
		case err := <-cerr:
			return &Msg{}, err
		}
	}
}

type Msg struct {
	Text string `json:"text"`
}

func (m *Msg) String() string {
	return m.Text
}

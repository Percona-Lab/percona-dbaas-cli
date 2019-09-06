package psmdb

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/pkg/errors"
)

func (p *PSMDB) Upgrade(apps map[string]string, ok chan<- string, msg chan<- dbaas.OutuputMsg, errc chan<- error) {
	acr, err := p.Cmd.GetObject(p.typ, p.name)
	if err != nil {
		errc <- errors.Wrap(err, "get config")
		return
	}

	err = p.upgrade(acr, apps)
	if err != nil {
		errc <- errors.Wrap(err, "upgrade cluster")
		return
	}

	cr, err := p.App()
	if err != nil {
		errc <- errors.Wrap(err, "get cr")
		return
	}
	err = p.Cmd.Apply(cr)
	if err != nil {
		errc <- errors.Wrap(err, "apply cr")
		return
	}

	time.Sleep(15 * time.Second)
	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		status, err := p.Cmd.GetObject(p.typ, p.name)
		if err != nil {
			errc <- errors.Wrap(err, "get cluster status")
			return
		}
		state, msgs, err := p.CheckStatus(status, make(map[string][]byte))
		if err != nil {
			errc <- errors.Wrap(err, "parse cluster status")
			return
		}

		switch state {
		case dbaas.ClusterStateReady:
			ok <- strings.Join(msgs, "\n")
			return
		case dbaas.ClusterStateError:
			errc <- errors.New(strings.Join(msgs, "\n"))
			return
		case dbaas.ClusterStateInit:
		}

		opLogsStream, err := p.Cmd.ReadOperatorLogs(p.name)
		if err != nil {
			errc <- errors.Wrap(err, "get operator logs")
			return
		}

		opLogs, err := p.CheckOperatorLogs(opLogsStream)
		if err != nil {
			errc <- errors.Wrap(err, "parse operator logs")
			return
		}

		for _, entry := range opLogs {
			msg <- entry
		}

		if tries >= p.Cmd.GetStatusMaxTries() {
			errc <- errors.Wrap(err, "unable to start cluster")
			return
		}

		tries++
	}
}

func (p *PSMDB) upgrade(crRaw []byte, newImages map[string]string) error {
	cr := &PerconaServerMongoDB{}
	err := json.Unmarshal(crRaw, cr)
	if err != nil {
		return errors.Wrap(err, "unmarshal current cr")
	}

	p.config.APIVersion = cr.APIVersion
	p.config.Kind = cr.Kind
	p.config.Name = cr.Name
	p.config.Spec = cr.Spec
	p.config.Status = cr.Status

	p.config.Upgrade(newImages)

	return nil
}

func (p *PSMDB) UpgradeOperator(newImage string, ok chan<- string, errc chan<- error) {
	if newImage == "" {
		return
	}

	for _, o := range p.Bundle(newImage) {
		if o.Kind == "Deployment" && o.Name == p.OperatorName() {
			err := p.Cmd.Apply(o.Data)
			if err != nil {
				errc <- errors.Wrap(err, "apply cr")
				return
			}

			time.Sleep(15 * time.Second)
			tries := 0
			tckr := time.NewTicker(500 * time.Millisecond)
			defer tckr.Stop()
			for range tckr.C {
				running, err := p.Cmd.GetOperatorPodStatus(p.OperatorName())
				if err != nil {
					errc <- errors.Wrap(err, "get operator status")
					return
				}
				if running {
					ok <- "Operator has been updated"
					return
				}

				if tries >= p.Cmd.GetStatusMaxTries() {
					errc <- errors.Wrap(err, "response limit has reached, unable get the success status from pod")
					return
				}

				tries++
			}
		}
	}
}

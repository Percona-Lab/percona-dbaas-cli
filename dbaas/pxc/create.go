package pxc

import (
	"strings"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/pkg/errors"
)

func (p *PXC) Create(ok chan<- string, msg chan<- dbaas.OutuputMsg, errc chan<- error) {
	p.Cmd.RunCmd(p.Cmd.ExecCommand, "create", "clusterrolebinding", "cluster-admin-binding", "--clusterrole=cluster-admin", "--user="+p.Cmd.OSUser())

	err := p.Cmd.ApplyBundles(p.Bundle(""))
	if err != nil {
		errc <- errors.Wrap(err, "apply bundles")
		return
	}

	ext, err := p.Cmd.IsObjExists(p.typ, p.Name())
	if err != nil {
		if strings.Contains(err.Error(), "error: the server doesn't have a resource type") ||
			strings.Contains(err.Error(), "Error from server (Forbidden):") {
			errc <- errors.Errorf(p.Cmd.GetOSRightsMsg(), p.Cmd.ExecCommand, p.Cmd.OSUser(), p.Cmd.ExecCommand, p.Cmd.OSAdminBundle(p.Bundle("")), p.Cmd.OSUser())
		}
		errc <- errors.Wrap(err, "check if cluster exists")
		return
	}

	if ext {
		errc <- dbaas.ErrAlreadyExists{Typ: p.typ, Cluster: p.Name()}
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

	// give a time for operator to start
	time.Sleep(1 * time.Minute)

	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		secrets, err := p.Cmd.GetSecrets(p.name)
		if err != nil {
			errc <- errors.Wrap(err, "get cluster secrets")
			return
		}
		status, err := p.Cmd.GetObject(p.typ, p.name)
		if err != nil {
			errc <- errors.Wrap(err, "get cluster status")
			return
		}
		state, msgs, err := p.CheckStatus(status, secrets)
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

		opLogsStream, err := p.Cmd.ReadOperatorLogs(p.OperatorName())
		if err != nil {
			// waiting for the operator to start
			if tries < p.Cmd.GetStatusMaxTries()/2 {
				continue
			}
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

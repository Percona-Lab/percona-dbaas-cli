package pxc

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/pkg/errors"
)

const updateMsg = `Update MySQL cluster.
 
PXC instances           | %v
ProxySQL instances      | %v
`

type UpdateMsg struct {
	Message           string `json:"message"`
	PXCInstances      int32  `json:"pxcInstances"`
	ProxySQLInstances int32  `json:"proxySQLInstances"`
}

func (p *PXC) edit(crRaw []byte, storage *dbaas.BackupStorageSpec) (string, error) {
	cr := &PerconaXtraDBCluster{}
	err := json.Unmarshal(crRaw, cr)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal current cr")
	}

	p.config.APIVersion = cr.APIVersion
	p.config.Kind = cr.Kind
	p.config.Name = cr.Name
	p.config.Finalizers = cr.Finalizers
	p.config.Spec = cr.Spec
	p.config.Status = cr.Status

	err = p.config.UpdateWith(p.ClusterConfig, storage)
	if err != nil {
		return "", errors.Wrap(err, "applay changes to cr")
	}

	if p.AnswerInJSON {
		updateJSONMsg := CreateMsg{
			Message:           "Update MySQL cluster",
			PXCInstances:      p.config.Spec.PXC.Size,
			ProxySQLInstances: p.config.Spec.ProxySQL.Size,
		}
		answer, err := json.Marshal(updateJSONMsg)
		if err != nil {
			return "", errors.Wrap(err, "marshal answer")
		}
		return string(answer), nil
	}

	return fmt.Sprintf(updateMsg, p.config.Spec.PXC.Size, p.config.Spec.ProxySQL.Size), nil
}

func (p PXC) Edit(storage *dbaas.BackupStorageSpec, ok chan<- string, msg chan<- dbaas.OutuputMsg, errc chan<- error) {
	acr, err := p.Cmd.GetObject(p.typ, p.name)
	if err != nil {
		errc <- errors.Wrap(err, "get config")
		return
	}

	_, err = p.edit(acr, storage)
	if err != nil {
		errc <- errors.Wrap(err, "update config")
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

		opLogsStream, err := p.Cmd.ReadOperatorLogs(p.OperatorName())
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

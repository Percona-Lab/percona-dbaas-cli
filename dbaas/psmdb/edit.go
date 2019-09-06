package psmdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/pkg/errors"
)

const updateMsg = `Update MongoDB cluster.
 
Replica Set Name        | %v
Replica Set Size        | %v
`

type UpdateMsg struct {
	Message        string `json:"message"`
	ReplicaSetName string `json:"replicaSetName"`
	ReplicaSetSize int32  `json:"replicaSetSize"`
}

func (p *PSMDB) edit(crRaw []byte, storage *dbaas.BackupStorageSpec) (string, error) {
	cr := &PerconaServerMongoDB{}
	err := json.Unmarshal(crRaw, cr)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal current cr")
	}

	p.config.APIVersion = cr.APIVersion
	p.config.Kind = cr.Kind
	p.config.Name = cr.Name
	p.config.Spec = cr.Spec
	p.config.Status = cr.Status

	err = p.config.UpdateWith(p.rsName, p.ClusterConfig, storage)
	if err != nil {
		return "", errors.Wrap(err, "apply changes to cr")
	}

	if p.AnswerInJSON {
		updateJSONMsg := UpdateMsg{
			Message:        "Update MongoDB cluster",
			ReplicaSetName: p.config.Spec.Replsets[0].Name,
			ReplicaSetSize: p.config.Spec.Replsets[0].Size,
		}
		answer, err := json.Marshal(updateJSONMsg)
		if err != nil {
			return "", errors.Wrap(err, "marshal answer")
		}
		return string(answer), nil
	}

	return fmt.Sprintf(updateMsg, p.config.Spec.Replsets[0].Name, p.config.Spec.Replsets[0].Size), nil
}

func (p PSMDB) Edit(storage *dbaas.BackupStorageSpec, ok chan<- string, msg chan<- dbaas.OutuputMsg, errc chan<- error) {
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

package pxc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/pkg/errors"
)

type Backup struct {
	name         string
	cluster      string
	valolume     string
	config       *PerconaXtraDBBackup
	opLogsLastTS float64
}

func NewBackup(cluster string) *Backup {
	return &Backup{
		cluster: cluster,
		config:  &PerconaXtraDBBackup{},
	}
}

func (b *Backup) Name() string {
	return b.name
}

func (b *Backup) Setup(storage string) {
	b.name = b.cluster + "." + time.Now().Format("2006-01-02.15-04-05")
	b.config.SetNew(b.name, b.cluster, storage)
}

func (b *Backup) CR() (string, error) {
	cr, err := json.Marshal(b.config)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(cr), nil
}

func (*Backup) OperatorName() string {
	return "percona-xtradb-cluster-operator"
}

func (b *Backup) CheckOperatorLogs(data []byte) ([]dbaas.OutuputMsg, error) {
	msgs := []dbaas.OutuputMsg{}

	lines := bytes.Split(data, []byte("\n"))
	for _, l := range lines {
		if len(l) == 0 {
			continue
		}

		entry := &operatorLog{}
		err := json.Unmarshal(l, entry)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal entry")
		}

		if entry.Controller != "perconaxtradbbackup-controller" {
			continue
		}

		// skips old entries
		if entry.TS <= b.opLogsLastTS {
			continue
		}

		b.opLogsLastTS = entry.TS

		if entry.Level != "error" {
			continue
		}

		cluster := ""
		s := strings.Split(entry.Request, "/")
		if len(s) == 2 {
			cluster = s[1]
		}

		if !strings.HasPrefix(cluster, b.cluster+".") {
			continue
		}

		msgs = append(msgs, alterOpError(entry))
	}

	return msgs, nil
}

const okmsgbcp = `
MySQL backup created successfully:
Name: %s
Destination: %s
`

func (b *Backup) CheckStatus(data []byte) (dbaas.ClusterState, []string, error) {
	st := &PerconaXtraDBBackup{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return dbaas.ClusterStateUnknown, nil, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.State {
	case BackupSucceeded:
		return dbaas.ClusterStateReady, []string{fmt.Sprintf(okmsgbcp, st.Name, st.Status.Destination)}, nil
	case BackupStarting, BackupRunning:
		return dbaas.ClusterStateInit, nil, nil
	case BackupFailed:
		return dbaas.ClusterStateError, []string{"backup attempt has failed"}, nil
	}

	return dbaas.ClusterStateInit, nil, nil
}

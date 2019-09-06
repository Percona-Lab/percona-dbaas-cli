// Copyright © 2019 Percona, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package psmdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

type Backup struct {
	name         string
	cluster      string
	config       *PerconaServerMongoDBBackup
	opLogsLastTS float64
	typ          string
	Cmd          *dbaas.Cmd
}

func NewBackup(cluster, envCrt string) (*Backup, error) {
	dbservice, err := dbaas.New(envCrt)
	if err != nil {
		return nil, err
	}
	return &Backup{
		cluster: cluster,
		config:  &PerconaServerMongoDBBackup{},
		typ:     "psmdb-backup",
		Cmd:     dbservice,
	}, nil
}

func (b *Backup) Name() string {
	return b.name
}

func (b *Backup) Setup(storage string) {
	b.name = time.Now().Format("20060102.150405") + "-" + dbaas.GenRandString(3)
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
	return "percona-server-mongodb-operator"
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

		if entry.Controller != "perconaservermongodbbackup-controller" {
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

		obj := ""
		s := strings.Split(entry.Request, "/")
		if len(s) == 2 {
			obj = s[1]
		}

		if obj != b.name {
			continue
		}

		msgs = append(msgs, alterOpError(entry))
	}

	return msgs, nil
}

const okmsgbcp = `
MongoDB backup created successfully:
Name: %s
Destination: %s
`

func (b *Backup) CheckStatus(data []byte) (dbaas.ClusterState, []string, error) {
	st := &PerconaServerMongoDBBackup{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return dbaas.ClusterStateUnknown, nil, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.State {
	case StateReady:
		return dbaas.ClusterStateReady, []string{fmt.Sprintf(okmsgbcp, st.Name, st.Status.Destination)}, nil
	case StateRequested:
		return dbaas.ClusterStateInit, nil, nil
	case StateRejected:
		return dbaas.ClusterStateError, []string{"backup attempt has failed"}, nil
	}

	return dbaas.ClusterStateInit, nil, nil
}

func (b *Backup) Create(ok chan<- string, msg chan<- dbaas.OutuputMsg, errc chan<- error) {
	cr, err := b.CR()
	if err != nil {
		errc <- errors.Wrap(err, "create cr")
		return
	}

	err = b.Cmd.Apply(cr)
	if err != nil {
		errc <- errors.Wrap(err, "apply cr")
		return
	}
	time.Sleep(1 * time.Minute)

	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		status, err := b.Cmd.GetObject(b.typ, b.name)
		if err != nil {
			errc <- errors.Wrap(err, "get cluster status")
			return
		}
		state, msgs, err := b.CheckStatus(status)
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

		opLogsStream, err := b.Cmd.ReadOperatorLogs(b.OperatorName())
		if err != nil {
			errc <- errors.Wrap(err, "get operator logs")
			return
		}

		opLogs, err := b.CheckOperatorLogs(opLogsStream)
		if err != nil {
			errc <- errors.Wrap(err, "parse operator logs")
			return
		}

		for _, entry := range opLogs {
			msg <- entry
		}

		if tries >= b.Cmd.GetStatusMaxTries() {
			errc <- errors.Wrap(err, "unable to create object")
			return
		}

		tries++
	}
}

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
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

type Restore struct {
	name         string
	cluster      string
	config       *PerconaServerMongoDBRestore
	opLogsLastTS float64
	typ          string
	Cmd          *dbaas.Cmd
}

func NewRestore(cluster, envCrt string) (*Restore, error) {
	dbservice, err := dbaas.New(envCrt)
	if err != nil {
		return nil, err
	}
	return &Restore{
		cluster: cluster,
		config:  &PerconaServerMongoDBRestore{},
		typ:     "psmdb-restore",
		Cmd:     dbservice,
	}, nil
}

func (b *Restore) Name() string {
	return b.name
}

func (b *Restore) Setup(backupName string) {
	b.name = time.Now().Format("20060102.150405") + "-" + dbaas.GenRandString(3)
	b.config.SetNew(b.name, b.cluster, backupName)
}

func (b *Restore) CR() (string, error) {
	cr, err := json.Marshal(b.config)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(cr), nil
}

func (*Restore) OperatorName() string {
	return "percona-server-mongodb-operator"
}

func (b *Restore) CheckOperatorLogs(data []byte) ([]dbaas.OutuputMsg, error) {
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

		if entry.Controller != "perconaservermongodbrestore-controller" {
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

type RestoreResponse struct {
	Message string `json:"message,omitempty"`
	Name    string `json:"name,omitempty"`
}

func (b *Restore) CheckStatus(data []byte) (dbaas.ClusterState, RestoreResponse, error) {
	st := &PerconaServerMongoDBRestore{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return dbaas.ClusterStateUnknown, RestoreResponse{}, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.State {
	case RestoreStateReady:
		return dbaas.ClusterStateReady, RestoreResponse{Message: "MongoDB backup restored successfully", Name: st.Name}, nil
	case RestoreStateRequested:
		return dbaas.ClusterStateInit, RestoreResponse{}, nil
	case RestoreStateRejected:
		return dbaas.ClusterStateError, RestoreResponse{Message: "restore attempt has failed"}, nil
	}

	return dbaas.ClusterStateInit, RestoreResponse{}, nil
}

func (b *Restore) Create(ok chan<- RestoreResponse, msg chan<- RestoreResponse, errc chan<- error) {
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
		state, resp, err := b.CheckStatus(status)
		if err != nil {
			errc <- errors.Wrap(err, "parse cluster status")
			return
		}

		switch state {
		case dbaas.ClusterStateReady:
			ok <- resp
			return
		case dbaas.ClusterStateError:
			errc <- errors.New(resp.Message)
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
			msg <- RestoreResponse{Message: entry.String()}
		}

		if tries >= b.Cmd.GetStatusMaxTries() {
			errc <- errors.Wrap(err, "unable to create object")
			return
		}

		tries++
	}
}

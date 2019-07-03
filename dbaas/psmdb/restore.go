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

type Restore struct {
	name         string
	cluster      string
	config       *PerconaServerMongoDBRestore
	opLogsLastTS float64
}

func NewRestore(cluster string) *Restore {
	return &Restore{
		cluster: cluster,
		config:  &PerconaServerMongoDBRestore{},
	}
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

const okmsgrestore = `
MongoDB backup restored successfully:
Name: %s
`

func (b *Restore) CheckStatus(data []byte) (dbaas.ClusterState, []string, error) {
	st := &PerconaServerMongoDBRestore{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return dbaas.ClusterStateUnknown, nil, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.State {
	case RestoreStateReady:
		return dbaas.ClusterStateReady, []string{fmt.Sprintf(okmsgrestore, st.Name)}, nil
	case RestoreStateRequested:
		return dbaas.ClusterStateInit, nil, nil
	case RestoreStateRejected:
		return dbaas.ClusterStateError, []string{"restore attempt has failed"}, nil
	}

	return dbaas.ClusterStateInit, nil, nil
}

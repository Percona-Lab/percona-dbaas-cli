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

package pxc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Percona-Lab/percona-dbaas-cli/operator/dbaas"
)

type Restore struct {
	name         string
	cluster      string
	config       *PerconaXtraDBClusterRestore
	opLogsLastTS float64
}

func NewRestore(cluster string) *Restore {
	return &Restore{
		cluster: cluster,
		config:  &PerconaXtraDBClusterRestore{},
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
	return "percona-xtradb-cluster-operator"
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

		if entry.Controller != "perconaxtradbclusterrestore-controller" {
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

type RestoreMsg struct {
	Message string `json:"message"`
	Name    string `json:"Name"`
}

func (r RestoreMsg) String() string {
	if len(r.Name) == 0 {
		return r.Message
	}
	okmsgbcp := `%s. Name: %s`
	return fmt.Sprintf(okmsgbcp, r.Message, r.Name)
}

func (b *Restore) CheckStatus(data []byte) (dbaas.ClusterState, dbaas.Msg, error) {
	st := &PerconaXtraDBClusterRestore{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return dbaas.ClusterStateUnknown, nil, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.State {
	case RestoreSucceeded:
		return dbaas.ClusterStateReady, RestoreMsg{
			Message: "MySQL backup restored successfully",
			Name:    st.Name,
		}, nil
	case RestoreFailed:
		return dbaas.ClusterStateError, RestoreMsg{Message: "restore attempt has failed"}, nil
	default:
		return dbaas.ClusterStateInit, nil, nil
	}
}

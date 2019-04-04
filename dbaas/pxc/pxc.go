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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

type Version string

var (
	Version030 Version = "0.3.0"
)

type PXC struct {
	name         string
	config       Config
	obj          dbaas.Objects
	dbpass       []byte
	opLogsLastTS float64
}

func New(name string, version Version) (*PXC, error) {
	pxc := &PXC{
		name:   name,
		obj:    objects[version],
		config: Config{ClusterName: name},
	}

	return pxc, nil
}

func (p PXC) Bundle() string {
	return p.obj.Bundle
}

func (p PXC) ClusterName() string {
	return p.name
}

func (p PXC) Secrets() (string, error) {
	pass := dbaas.GenSecrets(p.obj.Secrets.Keys)
	p.dbpass = pass["root"]
	pb64 := make(map[string]string, len(pass)+1)

	pb64["ClusterName"] = p.name

	for k, v := range pass {
		buf := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
		base64.StdEncoding.Encode(buf, v)
		pb64[k] = string(buf)
	}

	var buf bytes.Buffer
	err := p.obj.Secrets.Data.Execute(&buf, pb64)
	if err != nil {
		return "", errors.Wrap(err, "parse template:")
	}

	return buf.String(), nil
}

func (p PXC) App() (string, error) {
	var buf bytes.Buffer
	err := p.obj.CR.Execute(&buf, p.config)
	if err != nil {
		return "", errors.Wrap(err, "parse template:")
	}
	return buf.String(), nil
}

func (p *PXC) SetConfig(f *pflag.FlagSet) error {
	return p.config.Set(f)
}

type ClusterState string

const (
	ClusterStateInit  ClusterState = ""
	ClusterStateReady              = "ready"
)

// PerconaXtraDBClusterStatus defines the observed state of PerconaXtraDBCluster
type PerconaXtraDBClusterStatus struct {
	PXC      PodStatus    `json:"pxc,omitempty"`
	ProxySQL PodStatus    `json:"proxysql,omitempty"`
	Host     string       `json:"host,omitempty"`
	Status   ClusterState `json:"state,omitempty"`
}

type PodStatus struct {
	Size  int32 `json:"size,omitempty"`
	Ready int32 `json:"ready,omitempty"`
}

type k8sStatus struct {
	Status PerconaXtraDBClusterStatus
}

const okmsg = `
MySQL cluster started successfully, right endpoint for application:
Host: %s
Port: 3306
User: root
Pass: %s
`

func (p *PXC) CheckStatus(data []byte) (string, error) {
	st := &k8sStatus{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return "", errors.Wrap(err, "unmarshal status")
	}

	if st.Status.Status != ClusterStateReady {
		return "", dbaas.ErrorClusterNotReady
	}

	return fmt.Sprintf(okmsg, st.Status.Host, p.dbpass), nil
}

type operatorLog struct {
	Level   string  `json:"level"`
	TS      float64 `json:"ts"`
	Msg     string  `json:"msg"`
	Error   string  `json:"error"`
	Request string  `json:"Request"`
}

func (p *PXC) CheckOperatorLogs(data []byte) ([]dbaas.OutuputMsg, error) {
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

		// skips old entries
		if entry.TS <= p.opLogsLastTS {
			continue
		}

		p.opLogsLastTS = entry.TS

		if entry.Level != "error" {
			continue
		}

		cluster := ""
		s := strings.Split(entry.Request, "/")
		if len(s) == 2 {
			cluster = s[1]
		}

		if cluster != p.name {
			continue
		}

		msgs = append(msgs, alterOpError(entry))
	}

	return msgs, nil
}

func alterOpError(l *operatorLog) dbaas.OutuputMsg {
	if strings.Contains(l.Error, "the object has been modified; please apply your changes to the latest version and try again") {
		if i := strings.Index(l.Error, "Operation cannot be fulfilled on"); i >= 0 {
			return dbaas.OutuputMsgDebug(l.Error[i:])
		}
	}

	return dbaas.OutuputMsgError(l.Msg + ": " + l.Error)
}

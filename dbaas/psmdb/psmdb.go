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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

type Version string

var (
	Version100 Version = "1.0.0"
)

type PSMDB struct {
	name         string
	config       *PerconaServerMongoDB
	obj          dbaas.Objects
	dbpass       []byte
	opLogsLastTS float64
}

func New(name string, version Version) (*PSMDB, error) {
	pxc := &PSMDB{
		name:   name,
		obj:    objects[version],
		config: &PerconaServerMongoDB{},
	}

	return pxc, nil
}

func (p PSMDB) Bundle() []dbaas.BundleObject {
	return p.obj.Bundle
}

func (p PSMDB) Name() string {
	return p.name
}

func (p *PSMDB) Secrets() (string, error) {
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

func (p PSMDB) App() (string, error) {
	cr, err := json.Marshal(p.config)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(cr), nil
}

const createMsg = `Create MySQL cluster.
 
Replica Set Name        | %v
Replica Set Size        | %v
Storage                 | %v
`

func (p *PSMDB) Setup(f *pflag.FlagSet) (string, error) {
	err := p.config.SetNew(p.Name(), f)
	if err != nil {
		return "", errors.Wrap(err, "parse options")
	}

	storage, err := p.config.Spec.Replsets[0].VolumeSpec.PersistentVolumeClaim.Resources.Requests[corev1.ResourceStorage].MarshalJSON()
	if err != nil {
		return "", errors.Wrap(err, "marshal pxc volume requests")
	}

	return fmt.Sprintf(createMsg, p.config.Spec.Replsets[0].Name, p.config.Spec.Replsets[0].Size, string(storage)), nil
}

const updateMsg = `Update MySQL cluster.
 
Replica Set Name        | %v
Replica Set Size        | %v
`

func (p *PSMDB) Update(crRaw []byte, f *pflag.FlagSet) (string, error) {
	cr := &PerconaServerMongoDB{}
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

	err = p.config.UpdateWith(f)
	if err != nil {
		return "", errors.Wrap(err, "applay changes to cr")
	}

	return fmt.Sprintf(updateMsg, p.config.Spec.Replsets[0].Name, p.config.Spec.Replsets[0].Size), nil
}

func (p *PSMDB) OperatorName() string {
	return "percona-xtradb-cluster-operator"
}

func (p *PSMDB) NeedCertificates() []dbaas.TLSSecrets {
	return []dbaas.TLSSecrets{
		{
			SecretName: p.Name() + "-ssl",
			Hosts: []string{
				p.Name() + "-proxysql",
				"*." + p.Name() + "-proxysql",
			},
		},
		{
			SecretName: p.Name() + "-ssl-internal",
			Hosts: []string{
				p.Name() + "-pxc",
				"*." + p.Name() + "-pxc",
			},
		},
	}
}

type k8sStatus struct {
	Status PerconaServerMongoDBStatus
}

const okmsg = `
MongoDB cluster started successfully.

Enjoy!`

func (p *PSMDB) CheckStatus(data []byte) (dbaas.ClusterState, []string, error) {
	st := &k8sStatus{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return dbaas.ClusterStateUnknown, nil, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.Status {
	case AppStateReady:
		if len(p.dbpass) == 0 {
			p.dbpass = p.getDBPass()
		}
		return dbaas.ClusterStateReady, []string{okmsg}, nil
	case AppStateInit:
		return dbaas.ClusterStateInit, nil, nil
	case AppStateError:
		return dbaas.ClusterStateError, alterStatusMgs([]string{st.Status.Message}), nil
	}

	return dbaas.ClusterStateInit, nil, nil
}

func (p *PSMDB) getDBPass() []byte {
	sbytes, err := dbaas.GetObject("secret", p.Name()+"-secrets")
	if err != nil {
		return []byte("error:" + err.Error())
	}

	s := &corev1.Secret{}

	err = json.Unmarshal(sbytes, s)
	if err != nil {
		return []byte("error:" + err.Error())
	}

	pbytes, ok := s.Data["root"]
	if !ok {
		return []byte("<see cluster secrets>")
	}

	return pbytes
}

type operatorLog struct {
	Level      string  `json:"level"`
	TS         float64 `json:"ts"`
	Msg        string  `json:"msg"`
	Error      string  `json:"error"`
	Request    string  `json:"Request"`
	Controller string  `json:"Controller"`
}

func (p *PSMDB) CheckOperatorLogs(data []byte) ([]dbaas.OutuputMsg, error) {
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

		if entry.Controller != "perconaxtradbcluster-controller" {
			continue
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

func alterStatusMgs(msgs []string) []string {
	for i, msg := range msgs {
		msgs[i] = alterMessage(msg)
	}

	return msgs
}

func alterMessage(msg string) string {
	if strings.Contains(msg, "node(s) didn't match pod affinity/anti-affinity") {
		return "Cluster node(s) didn't satisfy pods [anti-]affinity rules. Try to change --anti-affinity-key parameter or add more nodes/change topology of your cluster."
	}

	if strings.Contains(msg, "Insufficient memory.") {
		return "Avaliable memory not enough to satisfy replica set request. Try to change --request-mem parameter or add more memmory to your cluster."
	}

	if strings.Contains(msg, "Insufficient cpu.") {
		return "Avaliable CPU not enough to satisfy replica set request. Try to change --request-cpu parameter or add more CPU to your cluster."
	}

	return msg
}

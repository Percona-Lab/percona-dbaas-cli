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

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

type Version string

var (
	Version100 Version = "1.0.0"
)

type PXC struct {
	name         string
	config       *PerconaXtraDBCluster
	obj          dbaas.Objects
	dbpass       []byte
	opLogsLastTS float64
}

func New(name string, version Version) *PXC {
	return &PXC{
		name:   name,
		obj:    objects[version],
		config: &PerconaXtraDBCluster{},
	}
}

func (p PXC) Bundle() []dbaas.BundleObject {
	return p.obj.Bundle
}

func (p PXC) Name() string {
	return p.name
}

func (p PXC) App() (string, error) {
	cr, err := json.Marshal(p.config)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(cr), nil
}

const createMsg = `Create MySQL cluster.
 
PXC instances           | %v
ProxySQL instances      | %v
Storage                 | %v
`

func (p *PXC) Setup(f *pflag.FlagSet, s3 *dbaas.BackupStorageSpec) (string, error) {
	err := p.config.SetNew(p.Name(), f, s3, dbaas.GetPlatformType())
	if err != nil {
		return "", errors.Wrap(err, "parse options")
	}

	storage, err := p.config.Spec.PXC.VolumeSpec.PersistentVolumeClaim.Resources.Requests[corev1.ResourceStorage].MarshalJSON()
	if err != nil {
		return "", errors.Wrap(err, "marshal pxc volume requests")
	}

	return fmt.Sprintf(createMsg, p.config.Spec.PXC.Size, p.config.Spec.ProxySQL.Size, string(storage)), nil
}

const updateMsg = `Update MySQL cluster.
 
PXC instances           | %v
ProxySQL instances      | %v
`

func (p *PXC) Edit(crRaw []byte, f *pflag.FlagSet, storage *dbaas.BackupStorageSpec) (string, error) {
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

	err = p.config.UpdateWith(f, storage)
	if err != nil {
		return "", errors.Wrap(err, "applay changes to cr")
	}

	return fmt.Sprintf(updateMsg, p.config.Spec.PXC.Size, p.config.Spec.ProxySQL.Size), nil
}

func (p *PXC) OperatorName() string {
	return "percona-xtradb-cluster-operator"
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

Enjoy!`

func (p *PXC) CheckStatus(data []byte, pass map[string][]byte) (dbaas.ClusterState, []string, error) {
	st := &k8sStatus{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return dbaas.ClusterStateUnknown, nil, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.Status {
	case AppStateReady:
		return dbaas.ClusterStateReady, []string{fmt.Sprintf(okmsg, st.Status.Host, pass["root"])}, nil
	case AppStateInit:
		return dbaas.ClusterStateInit, nil, nil
	case AppStateError:
		return dbaas.ClusterStateError, alterStatusMgs(st.Status.Messages), nil
	}

	return dbaas.ClusterStateInit, nil, nil
}

type operatorLog struct {
	Level      string  `json:"level"`
	TS         float64 `json:"ts"`
	Msg        string  `json:"msg"`
	Error      string  `json:"error"`
	Request    string  `json:"Request"`
	Controller string  `json:"Controller"`
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
	app := ""
	if i := strings.Index(msg, ":"); i >= 0 {
		app = msg[:i]
	}

	if strings.Contains(msg, "node(s) didn't match pod affinity/anti-affinity") {
		key := ""
		switch app {
		case "PXC":
			key = "--pxc-anti-affinity-key"
		case "ProxySQL":
			key = "--proxy-anti-affinity-key"
		}
		return fmt.Sprintf("Cluster node(s) didn't satisfy %s pods [anti-]affinity rules. Try to change %s parameter or add more nodes/change topology of your cluster.", app, key)
	}

	if strings.Contains(msg, "Insufficient memory.") {
		key := ""
		switch app {
		case "PXC":
			key = "--pxc-request-mem"
		case "ProxySQL":
			key = "--proxy-request-mem"
		}
		return fmt.Sprintf("Avaliable memory not enough to satisfy %s request. Try to change %s parameter or add more memmory to your cluster.", app, key)
	}

	if strings.Contains(msg, "Insufficient cpu.") {
		key := ""
		switch app {
		case "PXC":
			key = "--pxc-request-cpu"
		case "ProxySQL":
			key = "--proxy-request-cpu"
		}
		return fmt.Sprintf("Avaliable CPU not enough to satisfy %s request. Try to change %s parameter or add more CPU to your cluster.", app, key)
	}

	return msg
}

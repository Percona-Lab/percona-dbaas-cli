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

const (
	CurrentVersion Version = "default"

	defaultOperatorVersion = "percona/percona-xtradb-cluster-operator:1.1.0"
)

type PXC struct {
	name          string
	config        *PerconaXtraDBCluster
	obj           dbaas.Objects
	typ           string
	opLogsLastTS  float64
	AnswerInJSON  bool
	ClusterConfig ClusterConfig
	Cmd           *dbaas.Cmd
}

func New(name string, version Version, answerInJSON bool, labels, envCrt string) (*PXC, error) {
	config := &PerconaXtraDBCluster{}
	if len(labels) > 0 {
		config.ObjectMeta.Labels = make(map[string]string)
		keyValues := strings.Split(labels, ",")
		for index := range keyValues {
			itemSlice := strings.Split(keyValues[index], "=")
			config.ObjectMeta.Labels[itemSlice[0]] = itemSlice[1]
		}
	}
	dbservice, err := dbaas.New(envCrt)
	if err != nil {
		return nil, err
	}
	return &PXC{
		name:         name,
		obj:          Objects[version],
		config:       config,
		typ:          "pxc",
		AnswerInJSON: answerInJSON,
		Cmd:          dbservice,
	}, nil
}

func (p PXC) Bundle(operatorVersion string) []dbaas.BundleObject {
	if operatorVersion == "" {
		operatorVersion = defaultOperatorVersion
	}

	for i, o := range p.obj.Bundle {
		if o.Kind == "Deployment" && o.Name == p.OperatorName() {
			p.obj.Bundle[i].Data = strings.Replace(o.Data, "{{image}}", operatorVersion, -1)
		}
	}
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

type CreateMsg struct {
	Message           string `json:"message"`
	PXCInstances      int32  `json:"pxcInstances"`
	ProxySQLInstances int32  `json:"proxySQLInstances"`
	Storage           string `json:"storage"`
}

func (p *PXC) Setup(c ClusterConfig, s3 *dbaas.BackupStorageSpec, platform dbaas.PlatformType) (string, error) {
	err := p.config.SetNew(p.Name(), c, s3, platform)
	if err != nil {
		return "", errors.Wrap(err, "parse options")
	}

	storage, err := p.config.Spec.PXC.VolumeSpec.PersistentVolumeClaim.Resources.Requests[corev1.ResourceStorage].MarshalJSON()
	if err != nil {
		return "", errors.Wrap(err, "marshal pxc volume requests")
	}

	if p.AnswerInJSON {
		createJSONMsg := CreateMsg{
			Message:           "Create MySQL cluster",
			PXCInstances:      p.config.Spec.PXC.Size,
			ProxySQLInstances: p.config.Spec.ProxySQL.Size,
			Storage:           string(storage),
		}
		answer, err := json.Marshal(createJSONMsg)
		if err != nil {
			return "", errors.Wrap(err, "marshal answer")
		}
		return string(answer), nil
	}

	return fmt.Sprintf(createMsg, p.config.Spec.PXC.Size, p.config.Spec.ProxySQL.Size, string(storage)), nil
}

const operatorImage = "percona/percona-xtradb-cluster-operator:"

func (p *PXC) Images(ver string, f *pflag.FlagSet) (apps map[string]string, err error) {
	apps = make(map[string]string)

	if ver != "" {
		apps["pxc"] = operatorImage + ver + "-pxc"
		apps["proxysql"] = operatorImage + ver + "-proxysql"
		apps["backup"] = operatorImage + ver + "-backup"
	}

	pxc, err := f.GetString("database-image")
	if err != nil {
		return apps, errors.New("undefined `database-image`")
	}
	if pxc != "" {
		apps["pxc"] = pxc
	}

	proxysql, err := f.GetString("proxysql-image")
	if err != nil {
		return apps, errors.New("undefined `proxysql-image`")
	}
	if proxysql != "" {
		apps["proxysql"] = proxysql
	}

	backup, err := f.GetString("backup-image")
	if err != nil {
		return apps, errors.New("undefined `backup-image`")
	}
	if backup != "" {
		apps["backup"] = backup
	}

	return apps, nil
}

func (p *PXC) OperatorName() string {
	return "percona-xtradb-cluster-operator"
}

func (p *PXC) OperatorType() string {
	return "pxc"
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

type OkMsg struct {
	Message string `json:"message"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	User    string `json:"user"`
	Pass    string `json:"pass"`
}

func (p *PXC) CheckStatus(data []byte, pass map[string][]byte) (dbaas.ClusterState, []string, error) {
	st := &k8sStatus{}

	err := json.Unmarshal(data, st)
	if err != nil {
		return dbaas.ClusterStateUnknown, nil, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.Status {
	case AppStateReady:
		if p.AnswerInJSON {
			okJSONMsg := OkMsg{
				Message: "MySQL cluster started successfully",
				Host:    st.Status.Host,
				Port:    3306,
				User:    "root",
				Pass:    string(pass["root"]),
			}
			answer, err := json.Marshal(okJSONMsg)
			if err != nil {
				return dbaas.ClusterStateError, []string{}, errors.Wrap(err, "marshal answer")
			}
			return dbaas.ClusterStateReady, []string{string(answer)}, nil
		}
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

// JSONErrorMsg creates error messages in JSON format
func JSONErrorMsg(message string, err error) string {
	if err == nil {
		return fmt.Sprintf("\n{\"error\": \"%s\"}\n", message)
	}
	return fmt.Sprintf("\n{\"error\": \"%s: %v\"}\n", message, err)
}

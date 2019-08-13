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
	opLogsLastTS  float64
	AnswerInJSON  bool
	ClusterConfig ClusterConfig
}

func New(name string, version Version, answerInJSON bool, labels string) *PXC {
	config := &PerconaXtraDBCluster{}
	if len(labels) > 0 {
		config.ObjectMeta.Labels = make(map[string]string)
		keyValues := strings.Split(labels, ",")
		for index := range keyValues {
			itemSlice := strings.Split(keyValues[index], "=")
			config.ObjectMeta.Labels[itemSlice[0]] = itemSlice[1]
		}
	}
	return &PXC{
		name:         name,
		obj:          Objects[version],
		config:       config,
		AnswerInJSON: answerInJSON,
	}
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

const updateMsg = `Update MySQL cluster.
 
PXC instances           | %v
ProxySQL instances      | %v
`

type UpdateMsg struct {
	Message           string `json:"message"`
	PXCInstances      int32  `json:"pxcInstances"`
	ProxySQLInstances int32  `json:"proxySQLInstances"`
}

func (p *PXC) Edit(crRaw []byte, storage *dbaas.BackupStorageSpec) (string, error) {
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

	err = p.config.UpdateWith(p.ClusterConfig, storage)
	if err != nil {
		return "", errors.Wrap(err, "applay changes to cr")
	}

	if p.AnswerInJSON {
		updateJSONMsg := CreateMsg{
			Message:           "Update MySQL cluster",
			PXCInstances:      p.config.Spec.PXC.Size,
			ProxySQLInstances: p.config.Spec.ProxySQL.Size,
		}
		answer, err := json.Marshal(updateJSONMsg)
		if err != nil {
			return "", errors.Wrap(err, "marshal answer")
		}
		return string(answer), nil
	}

	return fmt.Sprintf(updateMsg, p.config.Spec.PXC.Size, p.config.Spec.ProxySQL.Size), nil
}

func (p *PXC) Upgrade(crRaw []byte, newImages map[string]string) error {
	cr := &PerconaXtraDBCluster{}
	err := json.Unmarshal(crRaw, cr)
	if err != nil {
		return errors.Wrap(err, "unmarshal current cr")
	}

	p.config.APIVersion = cr.APIVersion
	p.config.Kind = cr.Kind
	p.config.Name = cr.Name
	p.config.Finalizers = cr.Finalizers
	p.config.Spec = cr.Spec
	p.config.Status = cr.Status

	p.config.Upgrade(newImages)

	return nil
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

const describeMsg = `
Name:                                %v
Status:                              %v
Multi-AZ:                            %v
Labels:                              %v
 
PXC Count:                           %v
PXC Image:                           %v
PXC CPU Requests:                    %v
PXC Memory Requests:                 %v
PXC PodDisruptionBudget:             %v
PXC AntiAffinityTopologyKey:         %v
PXC StorageType:                     %v
PXC Allocated Storage:               %v
 
ProxySQL Count:                      %v
ProxySQL Image:                      %v
ProxySQL CPU Requests:               %v
ProxySQL Memory Requests:            %v
ProxySQL PodDisruptionBudget:        %v
ProxySQL AntiAffinityTopologyKey:    %v
ProxySQL StorageType:                %v
ProxySQL Allocated Storage:          %v
 
Backup Image:                        %v
Backup StorageType:                  %v
Backup Allocated Storage:            %v
Backup Schedule:                     %v
`

func (p *PXC) Describe(kubeInput []byte) (string, error) {
	cr := &PerconaXtraDBCluster{}
	err := json.Unmarshal([]byte(kubeInput), &cr)
	if err != nil {
		return "", errors.Wrapf(err, "json prase")
	}

	multiAz := "yes"
	noAzAffinityList := []string{"none", "hostname"}
	for _, arg := range noAzAffinityList {
		if *cr.Spec.PXC.Affinity.TopologyKey == arg {
			multiAz = "no"
		}
	}
	budgetPXC := map[string]string{"MinAvailable": "none", "MaxUnavailable": "none"}

	if cr.Spec.PXC.PodDisruptionBudget != nil && cr.Spec.PXC.PodDisruptionBudget.MinAvailable != nil {
		budgetPXC["MinAvailable"] = cr.Spec.PXC.PodDisruptionBudget.MinAvailable.String()
	}
	if cr.Spec.PXC.PodDisruptionBudget != nil && cr.Spec.PXC.PodDisruptionBudget.MaxUnavailable != nil {
		budgetPXC["MaxUnavailable"] = cr.Spec.PXC.PodDisruptionBudget.MaxUnavailable.String()
	}
	budgetSQL := map[string]string{"MinAvailable": "none", "MaxUnavailable": "none"}
	if cr.Spec.ProxySQL.PodDisruptionBudget != nil && cr.Spec.ProxySQL.PodDisruptionBudget.MinAvailable != nil {
		budgetSQL["MinAvailable"] = cr.Spec.ProxySQL.PodDisruptionBudget.MinAvailable.String()
	}
	if cr.Spec.ProxySQL.PodDisruptionBudget != nil && cr.Spec.ProxySQL.PodDisruptionBudget.MaxUnavailable != nil {
		budgetSQL["MaxUnavailable"] = cr.Spec.ProxySQL.PodDisruptionBudget.MaxUnavailable.String()
	}

	backupImage := "not set"
	backupSize := "not set"
	backupStorageClassName := "not set"
	backupSchedule := "not set"
	if cr.Spec.Backup != nil {
		backupImage = cr.Spec.Backup.Image

		if cr.Spec.Backup.Schedule != nil {
			backupSchedule = ""
			for schedule := range cr.Spec.Backup.Schedule {
				backupSchedule = backupSchedule + cr.Spec.Backup.Schedule[schedule].Name + ", "
			}
		}
		backupSchedule = strings.TrimRight(backupSchedule, ", ")
		for item := range cr.Spec.Backup.Storages {
			if cr.Spec.Backup.Storages[item].Type == "filesystem" {
				volume := cr.Spec.Backup.Storages[item]
				backupSizeBytes, err := volume.Volume.PersistentVolumeClaim.Resources.Requests["storage"].MarshalJSON()
				if err != nil {
					return "", err
				}
				backupSize = string(backupSizeBytes)
				backupStorageClassName = string(*volume.Volume.PersistentVolumeClaim.StorageClassName)
			}

		}
	}

	return fmt.Sprintf(describeMsg,
		cr.ObjectMeta.Name,
		cr.Status.Status,
		multiAz,
		dbaas.GetStringFromMap(cr.ObjectMeta.Labels),
		cr.Spec.PXC.Size,
		cr.Spec.PXC.Image,
		cr.Spec.PXC.Resources.Requests.CPU,
		cr.Spec.PXC.Resources.Requests.Memory,
		dbaas.GetStringFromMap(budgetPXC),
		*cr.Spec.PXC.Affinity.TopologyKey,
		cr.StorageClassesAllocated.PXC,
		cr.StorageSizeAllocated.PXC,
		cr.Spec.ProxySQL.Size,
		cr.Spec.ProxySQL.Image,
		cr.Spec.ProxySQL.Resources.Requests.CPU,
		cr.Spec.ProxySQL.Resources.Requests.Memory,
		dbaas.GetStringFromMap(budgetSQL),
		*cr.Spec.ProxySQL.Affinity.TopologyKey,
		cr.StorageClassesAllocated.ProxySQL,
		cr.StorageSizeAllocated.ProxySQL,
		backupImage,
		backupSize,
		backupStorageClassName,
		backupSchedule), nil
}

func (p *PXC) PodTypes() []string {
	return []string{"pxc", "proxysql"}
}

func (p *PXC) DataPodName(index int) string {
	return fmt.Sprintf("%s-pxc-%d", p.name, index)
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

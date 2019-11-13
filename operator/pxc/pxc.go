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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/Percona-Lab/percona-dbaas-cli/operator/dbaas"
)

type Version string

const (
	CurrentVersion Version = "default"

	defaultOperatorVersion = "percona/percona-xtradb-cluster-operator:1.1.0"
)

type pxc struct {
	name          string
	config        *PerconaXtraDBCluster
	obj           dbaas.Objects
	opLogsLastTS  float64
	ClusterConfig ClusterConfig
}

func new(name string, version Version, labels string) *pxc {
	config := &PerconaXtraDBCluster{}
	if len(labels) > 0 {
		config.ObjectMeta.Labels = make(map[string]string)
		keyValues := strings.Split(labels, ",")
		for index := range keyValues {
			itemSlice := strings.Split(keyValues[index], "=")
			config.ObjectMeta.Labels[itemSlice[0]] = itemSlice[1]
		}
	}
	return &pxc{
		name:   name,
		obj:    Objects[version],
		config: config,
	}
}

func (p pxc) bundle(operatorVersion string) []dbaas.BundleObject {
	if operatorVersion == "" {
		operatorVersion = defaultOperatorVersion
	}

	for i, o := range p.obj.Bundle {
		if o.Kind == "Deployment" && o.Name == p.operatorName() {
			p.obj.Bundle[i].Data = strings.Replace(o.Data, "{{image}}", operatorVersion, -1)
		}
	}
	return p.obj.Bundle
}

func (p pxc) getName() string {
	return p.name
}

func (p pxc) getCr() (string, error) {
	cr, err := json.Marshal(p.config)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(cr), nil
}

type k8sStatus struct {
	Status PerconaXtraDBClusterStatus
}

type Cluster struct {
	Host  string `json:"host,omitempty"`
	Port  int    `json:"port,omitempty"`
	User  string `json:"user,omitempty"`
	Pass  string `json:"pass,omitempty"`
	State string `json:"state,omitempty"`
}

func (c Cluster) String() string {
	stringMsg := `Host: %s, Port: 3306, User: root, Pass: %s`
	return fmt.Sprintf(stringMsg, c.Host, c.Pass)
}

func (p *pxc) CheckClusterStatus(data []byte, pass map[string][]byte) (Cluster, error) {
	st := &k8sStatus{}
	cluster := Cluster{}
	err := json.Unmarshal(data, st)
	if err != nil {
		return cluster, errors.Wrap(err, "unmarshal status")
	}

	switch st.Status.Status {
	case AppStateReady:
		cluster.Host = st.Status.Host
		cluster.Port = 3306
		cluster.User = "root"
		cluster.Pass = string(pass["root"])
		cluster.State = dbaas.ClusterStateReady
		return cluster, nil
	case AppStateInit:
		cluster.State = dbaas.ClusterStateInit
		return cluster, nil
	case AppStateError:
		cluster.State = dbaas.ClusterStateError
		return cluster, errors.New(st.Status.Messages[0])
	}
	return cluster, nil
}

func (p *pxc) Setup(c ClusterConfig, s3 *dbaas.BackupStorageSpec, platform dbaas.PlatformType) error {
	err := p.config.SetNew(p.name, c, s3, platform)
	if err != nil {
		return errors.Wrap(err, "parse options")
	}

	_, err = p.config.Spec.PXC.VolumeSpec.PersistentVolumeClaim.Resources.Requests[corev1.ResourceStorage].MarshalJSON()
	if err != nil {
		return errors.Wrap(err, "marshal pxc volume requests")
	}

	return nil
}

func (p *pxc) operatorName() string {
	return "percona-xtradb-cluster-operator"
}

func (p *pxc) OperatorType() string {
	return "pxc"
}

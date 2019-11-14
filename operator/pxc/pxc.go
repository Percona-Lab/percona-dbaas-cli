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
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/Percona-Lab/percona-dbaas-cli/operator/k8s"
)

const (
	defaultOperatorVersion = "percona/percona-xtradb-cluster-operator:1.1.0"
)

func (p PXC) bundle(operatorVersion string) []k8s.BundleObject {
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

func (p PXC) getCR() (string, error) {
	cr, err := json.Marshal(p.config)
	if err != nil {
		return "", errors.Wrap(err, "marshal cr template")
	}

	return string(cr), nil
}

type k8sStatus struct {
	Status PerconaXtraDBClusterStatus
}

func (p *PXC) Setup(c ClusterConfig, s3 *k8s.BackupStorageSpec, platform k8s.PlatformType) error {
	err := p.config.SetNew(c, s3, platform)
	if err != nil {
		return errors.Wrap(err, "parse options")
	}

	_, err = p.config.Spec.PXC.VolumeSpec.PersistentVolumeClaim.Resources.Requests[corev1.ResourceStorage].MarshalJSON()
	if err != nil {
		return errors.Wrap(err, "marshal pxc volume requests")
	}

	return nil
}

func (p *PXC) operatorName() string {
	return "percona-xtradb-cluster-operator"
}

func (p *PXC) OperatorType() string {
	return "pxc"
}

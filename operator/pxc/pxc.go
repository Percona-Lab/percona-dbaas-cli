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
	"strings"

	"github.com/pkg/errors"

	"github.com/Percona-Lab/percona-dbaas-cli/operator/k8s"
	"github.com/Percona-Lab/percona-dbaas-cli/operator/pxc/types/config"
	v100 "github.com/Percona-Lab/percona-dbaas-cli/operator/pxc/types/v100"
)

const (
	defaultOperatorVersion = "percona/percona-xtradb-cluster-operator:1.1.0"
)

type Version string

type k8sStatus struct {
	Status PerconaXtraDBClusterStatus
}

const (
	currentVersion Version = "default"
)

type VersionObject struct {
	k8s k8s.Objects
	pxc PXDBCluster
}

var objects map[Version]VersionObject

func init() {
	objects = make(map[Version]VersionObject)

	objects[currentVersion] = VersionObject{
		k8s: k8s.Objects{
			Bundle: v100.Bundle,
		},
		pxc: &v100.PerconaXtraDBCluster{},
	}
}

func (p PXC) bundle(v map[Version]VersionObject, operatorVersion string) []k8s.BundleObject {
	if operatorVersion == "" {
		operatorVersion = defaultOperatorVersion
	}

	for i, o := range v[currentVersion].k8s.Bundle {
		if o.Kind == "Deployment" && o.Name == p.operatorName() {
			v[currentVersion].k8s.Bundle[i].Data = strings.Replace(o.Data, "{{image}}", operatorVersion, -1)
		}
	}
	return v[currentVersion].k8s.Bundle
}

func (p PXC) getCR(cluster PXDBCluster) (string, error) {
	return cluster.GetCR()
}

func (p *PXC) setup(cluster PXDBCluster, c config.ClusterConfig, s3 *k8s.BackupStorageSpec, platform k8s.PlatformType) error {
	err := cluster.SetNew(c, s3, platform)
	if err != nil {
		return errors.Wrap(err, "parse options")
	}

	//_, err = p.config.Spec.PXC.VolumeSpec.PersistentVolumeClaim.Resources.Requests[corev1.ResourceStorage].MarshalJSON()
	err = cluster.MarshalRequests()
	if err != nil {
		return errors.Wrap(err, "marshal pxc volume requests")
	}

	return nil
}

func (p *PXC) operatorName() string {
	return "percona-xtradb-cluster-operator"
}

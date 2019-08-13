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

package dbaas

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func (p Cmd) List(typ string) (string, error) {
	out, err := p.runCmd("kubectl", "get", typ)
	if err != nil {
		return "", errors.Wrapf(err, "get cr: %s", out)
	}

	return string(out), nil
}

func (p Cmd) ListName(typ, name string) (string, error) {
	out, err := p.runCmd("kubectl", "get", typ, name)
	if err != nil {
		return "", errors.Wrapf(err, "get cr: %s", out)
	}

	return string(out), nil
}

func (p Cmd) Describe(app Deploy) (string, error) {
	out, err := p.runCmd("kubectl", "get", app.OperatorType(), app.Name(), "-o", "json")
	if err != nil {
		return "", errors.Wrapf(err, "get cr: %s", out)
	}

	mergedData := map[string]interface{}{}
	err = json.Unmarshal([]byte(out), &mergedData)
	if err != nil {
		return "", errors.Wrapf(err, "json prase")
	}

	cr, err := p.getPodInfo(app.DataPodName(0))
	if err != nil {
		return "", err
	}
	PVCs := map[string]string{}
	AllocatedStorage := map[string]string{}

	for item := range cr.Spec.Volumes {
		for pod := range app.PodTypes() {
			if cr.Spec.Volumes[item].Name == "datadir" || cr.Spec.Volumes[item].Name == "mongod-data" {
				pvcinfo, err := p.getPvcInfo(cr.Spec.Volumes[item].VolumeSource.PersistentVolumeClaim.ClaimName)
				if err != nil {
					return "", err
				}
				PVCs[app.PodTypes()[pod]] = *pvcinfo.Spec.StorageClassName

				qt, err := pvcinfo.Status.Capacity["storage"].MarshalJSON()
				if err != nil {
					return "", err
				}
				AllocatedStorage[app.PodTypes()[pod]] = string(qt)
			}
		}

	}
	mergedData["StorageClassesAllocated"] = PVCs
	mergedData["StorageSizeAllocated"] = AllocatedStorage

	out, err = json.Marshal(mergedData)
	if err != nil {
		return "", errors.Wrapf(err, "json pack")
	}
	return app.Describe(out)
}

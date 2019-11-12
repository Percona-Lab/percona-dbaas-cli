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
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
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

type MultiplePVCk8sOutput struct {
	Items []corev1.PersistentVolumeClaim `json:"items"`
}

func (p Cmd) Describe(app Deploy) (Msg, error) {
	out, err := p.runCmd("kubectl", "get", app.OperatorType(), app.Name(), "-o", "json")
	if err != nil {
		return nil, errors.Wrapf(err, "describe-db %s", out)
	}

	mergedData := map[string]interface{}{}
	err = json.Unmarshal([]byte(out), &mergedData)
	if err != nil {
		return nil, errors.Wrapf(err, "describe-db")
	}

	pvcList := &MultiplePVCk8sOutput{}
	pvcsJSON, err := p.runCmd("kubectl", "get", "pvc", fmt.Sprintf("--selector=app.kubernetes.io/instance=%s,app.kubernetes.io/managed-by=%s", app.Name(), app.OperatorName()), "-o", "json")
	if err != nil {
		return nil, errors.Wrapf(err, "describe-db")
	}
	err = json.Unmarshal([]byte(pvcsJSON), pvcList)
	if err != nil {
		return nil, errors.Wrapf(err, "describe-db")
	}
	PVCs := map[string]string{}
	AllocatedStorage := map[string]string{}

	for volume := range pvcList.Items {
		PVCs[pvcList.Items[volume].Labels["app.kubernetes.io/component"]] = *pvcList.Items[volume].Spec.StorageClassName
		qt, err := pvcList.Items[volume].Status.Capacity["storage"].MarshalJSON()
		if err != nil {
			return nil, errors.Wrapf(err, "describe-db")
		}
		AllocatedStorage[pvcList.Items[volume].Labels["app.kubernetes.io/component"]] = string(qt)
	}
	mergedData["StorageClassesAllocated"] = PVCs
	mergedData["StorageSizeAllocated"] = AllocatedStorage

	out, err = json.Marshal(mergedData)
	if err != nil {
		return nil, errors.Wrapf(err, "describe-db")
	}
	return app.Describe(out)
}

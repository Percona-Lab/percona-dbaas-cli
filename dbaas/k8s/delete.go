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

package k8s

import (
	"github.com/pkg/errors"
)

func (p Cmd) DeleteCluster(typ, operatorName, appName string, delPVC bool) error {
	err := p.delete(typ, appName)
	if err != nil {
		return errors.Wrap(err, "delete cluster")
	}
	if delPVC {
		err := p.deletePVC(operatorName, appName)
		if err != nil {
			return errors.Wrap(err, "delete cluster PVCs")
		}
	}

	return nil
}

func (p Cmd) delete(typ, name string) error {
	args := []string{
		"delete",
		typ,
		name,
	}
	if len(p.Namespace) > 0 {
		args = append(args, "-n", p.Namespace)
	}
	out, err := p.runCmd("kubectl", args...)

	return errors.Wrapf(err, "get cr: %s", out)

}

func (p Cmd) deletePVC(operatorName, appName string) error {
	out, err := p.runCmd("kubectl", "delete", "pvc",
		"-l", "app.kubernetes.io/part-of="+operatorName,
		"-l", "app.kubernetes.io/instance="+appName,
	)
	if err != nil {
		return errors.Wrapf(err, "get cr: %s", out)
	}

	return nil
}

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
	"os/exec"

	"github.com/pkg/errors"
)

func Delete(typ, name string, delPVC bool, ok chan<- string, errc chan<- error) {
	err := delete(typ, name)
	if err != nil {
		errc <- errors.Wrap(err, "delete cluster")
		return
	}
	if delPVC {
		err := deletePVC(typ, name)
		if err != nil {
			errc <- errors.Wrap(err, "delete cluster PVCs")
			return
		}
	}

	ok <- ""
}

func delete(typ, name string) error {
	out, err := exec.Command("kubectl", "delete", typ, name).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "get cr: %s", out)
	}

	return nil
}

func deletePVC(typ, name string) error {
	out, err := exec.Command("kubectl", "delete", "pvc",
		"-l", "app.kubernetes.io/part-of=percona-xtradb-cluster",
		"-l", "app.kubernetes.io/instance="+name,
	).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "get cr: %s", out)
	}

	return nil
}

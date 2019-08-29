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
	"github.com/pkg/errors"
)

func (p Cmd) Delete(typ string, app Deploy, delPVC bool, ok chan<- string, errc chan<- error) {
	err := p.delete(typ, app.Name())
	if err != nil {
		errc <- errors.Wrap(err, "delete cluster")
		return
	}
	if delPVC {
		err := p.deletePVC(app)
		if err != nil {
			errc <- errors.Wrap(err, "delete cluster PVCs")
			return
		}
	}

	ok <- ""
}

func (p Cmd) delete(typ, name string) error {
	if len(p.Namespace) > 0 {
		out, err := p.runCmd("kubectl", "delete", typ, name, "-n", p.Namespace)
		if err != nil {
			return errors.Wrapf(err, "get cr: %s", out)
		}
	} else {
		out, err := p.runCmd("kubectl", "delete", typ, name)
		if err != nil {
			return errors.Wrapf(err, "get cr: %s", out)
		}
	}

	return nil
}

func (p Cmd) deletePVC(app Deploy) error {
	out, err := p.runCmd("kubectl", "delete", "pvc",
		"-l", "app.kubernetes.io/part-of="+app.OperatorName(),
		"-l", "app.kubernetes.io/instance="+app.Name(),
	)
	if err != nil {
		return errors.Wrapf(err, "get cr: %s", out)
	}

	return nil
}

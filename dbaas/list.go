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

func List(typ string) (string, error) {
	out, err := exec.Command("kubectl", "get", typ).CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "get cr: %s", out)
	}

	return string(out), nil
}

func ListName(typ, name string) (string, error) {
	out, err := exec.Command("kubectl", "get", typ, name).CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "get cr: %s", out)
	}

	return string(out), nil
}

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
	"strings"

	"github.com/pkg/errors"
)

func (p Cmd) Upgrade(typ string, clusterName, cr string) error {
	ext, err := p.IsObjExists(typ, clusterName)
	if err != nil {
		if strings.Contains(err.Error(), "error: the server doesn't have a resource type") ||
			strings.Contains(err.Error(), "Error from server (Forbidden):") {
			//return errors.Errorf(osRightsMsg, p.execCommand, p.osUser(), p.execCommand, p.osUser())
			return err
		}
		return errors.Wrap(err, "check if cluster exists")
	}
	if !ext {
		return errors.New("cluster '" + clusterName + "' not exist")
	}
	err = p.apply(cr)
	if err != nil {
		return errors.Wrap(err, "apply cr")
	}

	return nil
}

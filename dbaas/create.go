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
	"encoding/base64"
	"math/rand"
	"os/exec"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Deploy interface {
	// Bundle returns crd, rbac and operator manifests
	Bundle() string
	// App returns application (custom resource) manifest
	App() (string, error)
	Secrets() (string, error)
}

type Objects struct {
	Bundle  string
	CR      *template.Template
	Secrets Secrets
}

type Secrets struct {
	Name string
	Data *template.Template
	Keys []string
	Rnd  *rand.Rand
}

func Create(app Deploy) error {
	err := apply(app.Bundle())
	if err != nil {
		return errors.Wrap(err, "apply bundle")
	}

	scrt, err := app.Secrets()
	if err != nil {
		return errors.Wrap(err, "get secrets")
	}
	err = apply(scrt)
	if err != nil {
		return errors.Wrap(err, "apply secrets")
	}

	cr, err := app.App()
	if err != nil {
		return errors.Wrap(err, "get cr")
	}
	err = apply(cr)
	if err != nil {
		return errors.Wrap(err, "apply cr")
	}

	return nil
}

func apply(yaml string) error {
	cmd := exec.Command("sh", "-c", "cat <<-EOF | kubectl apply -f -"+yaml+"\nEOF")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Errorf("%s", b)
	}

	return nil
}

func genPassB64() string {
	b := genPass()
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(b)))
	base64.StdEncoding.Encode(buf, b)
	return string(buf)
}

func genPass() []byte {
	pass := make([]byte, rand.Intn(5)+10)
	for i := 0; i < len(pass); i++ {
		pass[i] = '!' + byte(rand.Intn(94))
	}

	return pass
}

func GenSecrets(keys []string) map[string][]byte {
	pass := make(map[string][]byte, len(keys))
	for _, k := range keys {
		pass[k] = genPass()
	}

	return pass
}

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
	ClusterName() string

	CheckStatus(data []byte) (string, error)
	CheckOperatorLogs(data []byte) ([]string, error)
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

const getStatusMaxTries = 1200

var ErrorClusterNotReady = errors.New("not ready")

func Create(app Deploy, ok chan<- string, errc chan<- error) {
	err := apply(app.Bundle())
	if err != nil {
		errc <- errors.Wrap(err, "apply bundle")
		return
	}

	scrt, err := app.Secrets()
	if err != nil {
		errc <- errors.Wrap(err, "get secrets")
		return
	}
	err = apply(scrt)
	if err != nil {
		errc <- errors.Wrap(err, "apply secrets")
		return
	}

	cr, err := app.App()
	if err != nil {
		errc <- errors.Wrap(err, "get cr")
		return
	}
	err = apply(cr)
	if err != nil {
		errc <- errors.Wrap(err, "apply cr")
		return
	}

	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		status, err := getStatus(app.ClusterName())
		if err != nil {
			errc <- errors.Wrap(err, "get cluster status")
			return
		}
		resp, err := app.CheckStatus(status)
		if err == nil {
			ok <- resp
			return
		}
		if err != ErrorClusterNotReady {
			errc <- errors.Wrap(err, "parse cluster status")
			return
		}
		if tries >= getStatusMaxTries {
			errc <- errors.Wrap(err, "unable to run cluster")
			return
		}

		opLogsStream, err := readOperatorLogs("percona-xtradb-cluster-operator")
		if err != nil {
			errc <- errors.Wrap(err, "get operator logs")
			return
		}

		opLogs, err := app.CheckOperatorLogs(opLogsStream)
		if err != nil {
			errc <- errors.Wrap(err, "parse operator logs")
			return
		}

		for _, entry := range opLogs {
			errc <- errors.New(entry)
		}

		tries++
	}
}

func readOperatorLogs(operatorName string) ([]byte, error) {
	podName, err := exec.Command("kubectl", "get", "-l", "name="+operatorName, "-o", "jsonpath=\"{.items[0].metadata.name}\"").Output()
	if err != nil {
		return nil, errors.Wrap(err, "get operator pod name")
	}

	return exec.Command("kubectl", "logs", string(podName)).Output()
}

func getStatus(clusterName string) ([]byte, error) {
	cmd := exec.Command("kubectl", "get", "pxc/"+clusterName, "-o", "json")
	return cmd.Output()
}

func apply(yaml string) error {
	cmd := exec.Command("sh", "-c", "cat <<-EOF | kubectl apply -f -"+yaml+"\nEOF")
	b, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Errorf("%s", b)
	}

	return nil
}

var passsymbols = []byte("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func genPass() []byte {
	pass := make([]byte, rand.Intn(5)+16)
	for i := 0; i < len(pass); i++ {
		pass[i] = passsymbols[rand.Intn(len(passsymbols))]
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

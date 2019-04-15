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
	"fmt"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
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

	Name() string
	OperatorName() string

	CheckStatus(data []byte) (ClusterState, []string, error)
	CheckOperatorLogs(data []byte) ([]OutuputMsg, error)

	Update(crRaw []byte, f *pflag.FlagSet) (string, error)
}

type ClusterState string

const (
	ClusterStateUnknown ClusterState = "unknown"
	ClusterStateInit                 = "initializing"
	ClusterStateReady                = "ready"
	ClusterStateError                = "error"
)

type Objects struct {
	Bundle  string
	Secrets Secrets
}

type Secrets struct {
	Name string
	Data *template.Template
	Keys []string
	Rnd  *rand.Rand
}

const getStatusMaxTries = 1200

type ErrAlreadyExists struct {
	Typ     string
	Cluster string
}

func (e ErrAlreadyExists) Error() string {
	return fmt.Sprintf("cluster %s/%s already exists", e.Typ, e.Cluster)
}

func Create(typ string, app Deploy, ok chan<- string, msg chan<- OutuputMsg, errc chan<- error) {
	err := apply(app.Bundle())
	if err != nil {
		errc <- errors.Wrap(err, "apply bundle")
		return
	}

	ext, err := IsCRexists(typ, app.Name())
	if err != nil {
		errc <- errors.Wrap(err, "check if cluster exists")
		return
	}

	if ext {
		errc <- ErrAlreadyExists{Typ: typ, Cluster: app.Name()}
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

	// give a time for operator to start
	time.Sleep(1 * time.Minute)

	tries := 0
	tckr := time.NewTicker(500 * time.Millisecond)
	defer tckr.Stop()
	for range tckr.C {
		status, err := getCR(typ, app.Name())
		if err != nil {
			errc <- errors.Wrap(err, "get cluster status")
			return
		}
		state, msgs, err := app.CheckStatus(status)
		if err != nil {
			errc <- errors.Wrap(err, "parse cluster status")
			return
		}

		switch state {
		case ClusterStateReady:
			ok <- strings.Join(msgs, "\n")
			return
		case ClusterStateError:
			errc <- errors.New(strings.Join(msgs, "\n"))
			return
		case ClusterStateInit:
		}

		opLogsStream, err := readOperatorLogs(app.OperatorName())
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
			msg <- entry
		}

		if tries >= getStatusMaxTries {
			errc <- errors.Wrap(err, "unable to start cluster")
			return
		}

		tries++
	}
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

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
	Bundle() []BundleObject
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

type BundleObject struct {
	Kind string
	Name string
	Data string
}

type Objects struct {
	Bundle  []BundleObject
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

const osRightsMsg = `Not enough rights to pre-setup cluster.
Try to login under the privileged user run one of the commands listed below and then call "percona-dbaas pxc create ..." again

1) %s create clusterrolebinding cluster-admin-binding --clusterrole=cluster-admin --user=%s

or

2) cat <<-EOF | %s apply -f -
%s
EOF

oc create clusterrole pxc-admin --verb="*" --resource=perconaxtradbclusters.pxc.percona.com,perconaxtradbbackups.pxc.percona.com,perconaxtradbbackuprestores.pxc.percona.com,perconaxtradbbackuprestores.pxc.percona.com/status,issuers.certmanager.k8s.io,certificates.certmanager.k8s.io
oc adm policy add-cluster-role-to-user pxc-admin %s
`

func Create(typ string, app Deploy, ok chan<- string, msg chan<- OutuputMsg, errc chan<- error) {
	runCmd(execCommand, "create", "clusterrolebinding", "cluster-admin-binding", "--clusterrole=cluster-admin", "--user="+osUser())
	runCmd(execCommand, "create", "clusterrole", "pxc-admin --verb=\"*\" --resource=perconaxtradbclusters.pxc.percona.com,perconaxtradbbackups.pxc.percona.com,perconaxtradbbackuprestores.pxc.percona.com,perconaxtradbbackuprestores.pxc.percona.com/status,issuers.certmanager.k8s.io,certificates.certmanager.k8s.io")
	runCmd("oc", "adm", "policy", "add-cluster-role-to-user", "pxc-admin", osUser())

	err := applyBundles(app.Bundle())
	if err != nil {
		errc <- errors.Wrap(err, "apply bundles")
		return
	}

	ext, err := IsObjExists(typ, app.Name())
	if err != nil {
		if strings.Contains(err.Error(), "error: the server doesn't have a resource type") ||
			strings.Contains(err.Error(), "Error from server (Forbidden):") {
			errc <- errors.Errorf(osRightsMsg, execCommand, osUser(), execCommand, osAdminBundle(app.Bundle()), osUser())
		}
		errc <- errors.Wrap(err, "check if cluster exists")
		return
	}

	if ext {
		errc <- ErrAlreadyExists{Typ: typ, Cluster: app.Name()}
		return
	}

	secExt, err := IsObjExists("secret", app.Name()+"-secrets")
	if err != nil {
		errc <- errors.Wrap(err, "check if cluster secrets exists")
		return
	}

	if !secExt {
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
		status, err := GetObject(typ, app.Name())
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
			// waiting for the operator to start
			if tries < getStatusMaxTries/2 {
				continue
			}
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

func osAdminBundle(bs []BundleObject) string {
	objs := []string{}
	for _, b := range bs {
		switch b.Kind {
		case "CustomResourceDefinition", "Role", "RoleBinding":
			objs = append(objs, strings.TrimSpace(b.Data))
		}
	}

	return strings.Join(objs, "\n---\n")
}

func applyBundles(bs []BundleObject) error {
	for _, b := range bs {
		err := apply(b.Data)
		if err != nil {
			switch b.Kind {
			case "CustomResourceDefinition", "Role":
				if strings.Contains(err.Error(), fmt.Sprintf(`"%s" is forbidden:`, b.Name)) {
					continue
				}
			case "RoleBinding":
				if strings.Contains(err.Error(), "Error from server (NotFound)") {
					continue
				}
			}
			return errors.Wrapf(err, "apply %s/%s", b.Kind, b.Name)
		}
	}

	return nil
}

func osUser() string {
	ret := "<Your Opeshift User>"
	s, err := runCmd("oc", "whoami")
	if err != nil {
		u, err := gkeUser()
		if err != nil {
			return ret
		}
		return u
	}

	if len(s) > 0 {
		return strings.TrimSpace(string(s))
	}

	return ret
}

func gkeUser() (string, error) {
	s, err := runCmd("gcloud", "config", "get-value", "core/account")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(s)), nil
}

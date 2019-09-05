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
	"math/rand"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Deploy interface {
	// Bundle returns crd, rbac and operator manifests
	Bundle(operatorVersion string) []BundleObject
	// App returns application (custom resource) manifest
	App() (string, error)

	Name() string
	OperatorName() string
	OperatorType() string

	CheckStatus(data []byte, secrets map[string][]byte) (ClusterState, []string, error)
	CheckOperatorLogs(data []byte) ([]OutuputMsg, error)

	Edit(crRaw []byte, storage *BackupStorageSpec) (string, error)
	Upgrade(crRaw []byte, newImages map[string]string) error
	Describe(crRaw []byte) (string, error)
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
	Bundle []BundleObject
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

oc create clusterrole pxc-admin --verb="*" --resource=perconaxtradbclusters.pxc.percona.com
oc adm policy add-cluster-role-to-user pxc-admin %s
`

func (p Cmd) GetOSRightsMsg() string {
	return osRightsMsg
}

func (p Cmd) GetStatusMaxTries() int {
	return getStatusMaxTries
}

func (p Cmd) Create(typ string, app Deploy, ok chan<- string, msg chan<- OutuputMsg, errc chan<- error) {
	p.RunCmd(p.ExecCommand, "create", "clusterrolebinding", "cluster-admin-binding", "--clusterrole=cluster-admin", "--user="+p.OSUser())

	err := p.ApplyBundles(app.Bundle(""))
	if err != nil {
		errc <- errors.Wrap(err, "apply bundles")
		return
	}

	ext, err := p.IsObjExists(typ, app.Name())
	if err != nil {
		if strings.Contains(err.Error(), "error: the server doesn't have a resource type") ||
			strings.Contains(err.Error(), "Error from server (Forbidden):") {
			errc <- errors.Errorf(osRightsMsg, p.ExecCommand, p.OSUser(), p.ExecCommand, p.OSAdminBundle(app.Bundle("")), p.OSUser())
		}
		errc <- errors.Wrap(err, "check if cluster exists")
		return
	}

	if ext {
		errc <- ErrAlreadyExists{Typ: typ, Cluster: app.Name()}
		return
	}

	cr, err := app.App()
	if err != nil {
		errc <- errors.Wrap(err, "get cr")
		return
	}
	err = p.Apply(cr)
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
		secrets, err := p.GetSecrets(app.Name())
		if err != nil {
			errc <- errors.Wrap(err, "get cluster secrets")
			return
		}
		status, err := p.GetObject(typ, app.Name())
		if err != nil {
			errc <- errors.Wrap(err, "get cluster status")
			return
		}
		state, msgs, err := app.CheckStatus(status, secrets)
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

		opLogsStream, err := p.ReadOperatorLogs(app.OperatorName())
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

// CreateSecret creates k8s secret object with the given name and data
func (p Cmd) CreateSecret(name string, data map[string][]byte) error {
	s := corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Data: data,
		Type: corev1.SecretTypeOpaque,
	}

	sj, err := json.Marshal(s)
	if err != nil {
		errors.Wrap(err, "json marshal")
	}

	return errors.WithMessage(p.Apply(string(sj)), "apply")
}

func (p Cmd) OSAdminBundle(bs []BundleObject) string {
	objs := []string{}
	for _, b := range bs {
		switch b.Kind {
		case "CustomResourceDefinition", "Role", "RoleBinding":
			objs = append(objs, strings.TrimSpace(b.Data))
		}
	}

	return strings.Join(objs, "\n---\n")
}

func (p Cmd) ApplyBundles(bs []BundleObject) error {
	for _, b := range bs {
		err := p.Apply(b.Data)
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

func (p Cmd) OSUser() string {
	ret := "<Your Opeshift User>"
	s, err := p.RunCmd("oc", "whoami")
	if err != nil {
		u, err := p.GKEUser()
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

func (p Cmd) GKEUser() (string, error) {
	s, err := p.RunCmd("gcloud", "config", "get-value", "core/account")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(s)), nil
}

func (p Cmd) GetSecrets(appName string) (map[string][]byte, error) {
	data, err := p.GetObject("secrets", appName+"-secrets")
	if err != nil {
		return nil, errors.Wrap(err, "get object")
	}

	secretsObj := &corev1.Secret{}
	err = json.Unmarshal(data, secretsObj)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}

	return secretsObj.Data, nil
}

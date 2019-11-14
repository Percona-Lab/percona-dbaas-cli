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
	bundle(operatorVersion string) []BundleObject
	// App returns application (custom resource) manifest
	getCR() (string, error)

	OperatorType() string
}

type Msg interface {
	String() string
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

func (p Cmd) CreateCluster(typ, operatorVersion, clusterName, cr string, bundle []BundleObject) error {
	p.runCmd(p.execCommand, "create", "clusterrolebinding", "cluster-admin-binding", "--clusterrole=cluster-admin", "--user="+p.osUser())

	err := p.applyBundles(bundle)
	if err != nil {
		return errors.Wrap(err, "apply bundles")
	}

	ext, err := p.IsObjExists(typ, clusterName)
	if err != nil {
		if strings.Contains(err.Error(), "error: the server doesn't have a resource type") ||
			strings.Contains(err.Error(), "Error from server (Forbidden):") {
			return errors.Errorf(osRightsMsg, p.execCommand, p.osUser(), p.execCommand, osAdminBundle(bundle), p.osUser())
		}
		return errors.Wrap(err, "check if cluster exists")
	}
	if ext {
		return ErrAlreadyExists{Typ: typ, Cluster: clusterName}
	}

	err = p.apply(cr)
	if err != nil {
		return errors.Wrap(err, "apply cr")
	}

	return nil
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

	return errors.WithMessage(p.apply(string(sj)), "apply")
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

func (p Cmd) applyBundles(bs []BundleObject) error {
	for _, b := range bs {
		err := p.apply(b.Data)
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

func (p Cmd) osUser() string {
	ret := "<Your Opeshift User>"
	s, err := p.runCmd("oc", "whoami")
	if err != nil {
		u, err := p.gkeUser()
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

func (p Cmd) gkeUser() (string, error) {
	s, err := p.runCmd("gcloud", "config", "get-value", "core/account")
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

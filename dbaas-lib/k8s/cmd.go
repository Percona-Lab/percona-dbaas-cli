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
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	ErrOutOfMemory = errors.New("out of memory")
	ErrNotFound    = errors.New("not found")
)

type PlatformType string

const (
	PlatformKubernetes PlatformType = "kubernetes"
	PlatformMinikube   PlatformType = "minikube"
	PlatformOpenshift  PlatformType = "openshift"
	PlatformMinishift  PlatformType = "minishift"
)

type Cmd struct {
	environment string
	Namespace   string
	execCommand string
}

type ErrCmdRun struct {
	cmd    string
	args   []string
	output []byte
}

type Clusters struct {
	Items []interface{} `json:"items"`
}

type Pods struct {
	Items []corev1.Pod `json:"items"`
}

func (e ErrCmdRun) Error() string {
	return fmt.Sprintf("failed to run `%s %s`, output: %s", e.cmd, strings.Join(e.args, " "), e.output)
}

func New(environment string) (*Cmd, error) {
	execCommand := k8sExecDefault
	if _, err := exec.LookPath(execCommand); err != nil {
		execCommand = k8sExecCustom
		if _, err := exec.LookPath(execCommand); err != nil {
			return nil, fmt.Errorf("unable to find neither '%s' nor '%s' exec files", k8sExecDefault, k8sExecCustom)
		}
	}

	targetKubeConfig := os.Getenv("HOME") + "/.percona/" + environment + "/kubeconfig"
	if len(environment) > 0 {
		if _, err := exec.LookPath(targetKubeConfig); err != nil {
			return &Cmd{
				environment: targetKubeConfig,
			}, nil

		}
		files, err := ioutil.ReadDir(os.Getenv("HOME") + "/.percona/")
		if err != nil {
			return nil, fmt.Errorf("can't read the content of ~/.percona: %v", err)
		}
		var dirs []string
		for _, file := range files {
			if _, err := exec.LookPath(os.Getenv("HOME") + "/.percona/" + file.Name() + "/kubeconfig"); err != nil && file.IsDir() {
				dirs = append(dirs, file.Name())
			}
		}

		return nil, fmt.Errorf("can't find the requested env. Please use one of ther following: %v", dirs)

	}
	return &Cmd{
		environment: environment,
		execCommand: execCommand,
	}, nil
}

func (p Cmd) runCmd(cmd string, args ...string) ([]byte, error) {
	o, err := p.runNTimes(3, cmd, args...)
	if err != nil {

		return nil, ErrCmdRun{cmd: cmd, args: args, output: o}
	}

	return o, nil
}

func (p Cmd) runNTimes(n int, cmd string, args ...string) (o []byte, err error) {
	for i := 1; i <= n; i++ {
		cli := exec.Command(cmd, args...)
		cli.Env = os.Environ()
		if len(p.environment) > 0 {
			cli.Env = append(cli.Env, "KUBECONFIG="+p.environment)
		}
		o, err = cli.CombinedOutput()
		if err != nil {
			if strings.Contains(string(o), "Unable to connect to the server") && i < n {
				continue
			}
		}
		break
	}

	return o, err
}

func (p Cmd) readOperatorLogs(operatorName string) ([]byte, error) {
	return p.runCmd(p.execCommand, "logs", "-l", "name="+operatorName)
}

func (p Cmd) GetObjectsElement(typ, name, jsonPath string) ([]byte, error) {
	args := []string{"get", typ, name, "-o=jsonpath={" + jsonPath + "}"}
	if len(p.Namespace) > 0 {
		args = append(args, []string{"-n", p.Namespace}...)
	}
	data, err := p.runCmd(p.execCommand, args...)
	if err != nil && strings.Contains(err.Error(), "not found") {
		err = ErrNotFound
	} else if err != nil && strings.Contains(err.Error(), "doesn't have a resource") {
		err = ErrNotFound
	}
	if strings.Contains(string(data), "not found") {
		err = ErrNotFound
	}

	return data, err
}

func (p Cmd) GetObject(typ, name string) (objData []byte, err error) {
	if len(p.Namespace) > 0 {
		objData, err = p.runCmd(p.execCommand, "get", typ+"/"+name, "-n", p.Namespace, "-o", "json")
	} else {
		objData, err = p.runCmd(p.execCommand, "get", typ+"/"+name, "-o", "json")
	}
	if err != nil && strings.Contains(err.Error(), "Not found") {
		err = ErrNotFound
	}

	return
}

func (p Cmd) GetObjects(typ string) ([]byte, error) {
	args := []string{"get", typ, "-o", "json"}
	if len(p.Namespace) > 0 {
		args = append(args, []string{"-n", p.Namespace}...)
	}
	data, err := p.runCmd(p.execCommand, args...)
	if err != nil && strings.Contains(err.Error(), "not found") {
		err = ErrNotFound
	}
	if err != nil && strings.Contains(string(data), "not found") {
		err = ErrNotFound
	} else if err != nil && strings.Contains(err.Error(), "doesn't have a resource") {
		err = ErrNotFound
	}

	if strings.Contains(string(data), "No resources found") {
		err = ErrNotFound
	}

	return data, err
}

func (p Cmd) DeleteObject(typ, name string) error {
	if len(p.Namespace) > 0 {
		_, err := p.runCmd(p.execCommand, "delete", typ+"/"+name, "-n", p.Namespace)
		return err

	}
	_, err := p.runCmd(p.execCommand, "delete", typ+"/"+name)

	return err
}

func (p Cmd) apply(k8sObj string) error {
	namespace := ""
	if len(p.Namespace) > 0 {
		namespace = "-n=" + p.Namespace
	}
	fileName := os.TempDir() + "percona-dbaascli-temp-cr.json"
	f, err := os.Create(fileName)
	if err != nil {
		return errors.Wrap(err, "create cr file")
	}
	defer os.Remove(fileName)

	_, err = f.Write([]byte(k8sObj))
	if err != nil {
		return errors.Wrap(err, "write cr file")
	}
	if len(namespace) > 0 {
		_, err = p.runCmd(p.execCommand, "apply", namespace, "-f", fileName)
	} else {
		_, err = p.runCmd(p.execCommand, "apply", "-f", fileName)
	}
	if err != nil {
		return err
	}

	return nil
}

func (p Cmd) GetCurrentNamespace() (string, error) {
	o, err := p.runCmd(p.execCommand, "config", "view", "--minify", "--output", "jsonpath={..namespace}")
	if err != nil {
		return "", err
	}
	return string(o), err
}

func (p Cmd) Annotate(resource, clusterName, annotName, instance string) error {
	_, err := p.runCmd(p.execCommand, "annotate", resource, clusterName, annotName+"="+instance, "--overwrite=true")

	return err
}

func (p Cmd) IsObjExists(typ, name string) (bool, error) {
	switch typ {
	case "pxc":
		typ = "perconaxtradbcluster.pxc.percona.com"
	case "psmdb":
		typ = "perconaservermongodb.psmdb.percona.com"
	case "pxc-backup":
		typ = "perconaxtradbclusterbackup.pxc.percona.com"
	case "psmdb-backup":
		typ = "perconaservermongodbbackup.psmdb.percona.com"
	}

	out, err := p.runCmd(p.execCommand, "get", typ, name, "-o", "name")
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return false, errors.Wrapf(err, "get cr: %s", out)
	}

	return strings.TrimSpace(string(out)) == typ+"/"+name, nil
}

func (p Cmd) Instances(typ string) ([]string, error) {
	out, err := p.runCmd(p.execCommand, "get", typ, "-o", "name")
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return nil, errors.Wrapf(err, "get objects: %s", out)
	}

	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

func (p Cmd) GetServiceBrokerInstances(typ string) ([]byte, error) {
	out, err := p.runCmd(p.execCommand, "get", typ, "-o", "jsonpath='{.items..metadata.annotations.broker-instance}'")
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return nil, errors.Wrapf(err, "get objects: %s", out)
	}

	return out, nil
}

const genSymbols = "abcdefghijklmnopqrstuvwxyz1234567890"

// GenRandString generates a k8s-name legitimate string of given length
func GenRandString(ln int) string {
	b := make([]byte, ln)
	for i := range b {
		b[i] = genSymbols[rand.Intn(len(genSymbols))]
	}

	return string(b)
}

// GetPlatformType is for determine and return platform type
func (p Cmd) GetPlatformType() PlatformType {
	if p.checkMinikube() {
		return PlatformMinikube
	}

	if p.checkMinishift() {
		return PlatformMinishift
	}

	if p.checkOpenshift() {
		return PlatformOpenshift
	}

	return PlatformKubernetes
}

func (p Cmd) checkMinikube() bool {
	output, err := p.runCmd(p.execCommand, "get", "storageclass", "-o", "jsonpath='{.items..provisioner}'")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "k8s.io/minikube-hostpath")
}

func (p Cmd) checkMinishift() bool {
	output, err := p.runCmd(p.execCommand, "get", "pods", "master-etcd-localhost", "-n", "kube-system", "-o", "jsonpath='{.spec.volumes..path}'")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "minishift")
}

func (p Cmd) checkOpenshift() bool {
	output, err := p.runCmd(p.execCommand, "api-versions")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "openshift")
}

func GetStringFromMap(input map[string]string) string {
	b := new(bytes.Buffer)
	for key, value := range input {
		fmt.Fprintf(b, "%s=\"%s\", ", key, value)
	}
	if len(b.String()) > 0 {
		return strings.TrimSuffix(b.String(), ", ")
	}
	return "none"
}

func (p Cmd) GetObjectByLables(typ, lables string) ([]byte, error) {
	return p.runCmd(p.execCommand, "get", typ, "-l", lables, "-o", "json")
}

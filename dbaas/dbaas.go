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
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	execCommand = k8sExecDefault
	if _, err := exec.LookPath(execCommand); err != nil {
		execCommand = k8sExecCustom
		if _, err := exec.LookPath(execCommand); err != nil {
			panic(fmt.Sprintf("Unable to find neither '%s' nor '%s' exec files", k8sExecDefault, k8sExecCustom))
		}
	}
}

type PlatformType string

const (
	PlatformKubernetes PlatformType = "kubernetes"
	PlatformMinikube   PlatformType = "minikube"
	PlatformOpenshift  PlatformType = "openshift"
	PlatformMinishift  PlatformType = "minishift"
)

var execCommand string

type DBAAS struct {
	environment string
}

type ErrCmdRun struct {
	cmd    string
	args   []string
	output []byte
}

type ClusterConfig struct {
	PXC      Spec
	ProxySQL Spec
	PSMDB    Spec
	S3       S3StorageConfig
}

type Spec struct {
	StorageSize     string
	StorageClass    string
	Instances       int32
	RequestCPU      string
	RequestMem      string
	AntiAffinityKey string
}

func (e ErrCmdRun) Error() string {
	return fmt.Sprintf("failed to run `%s %s`, output: %s", e.cmd, strings.Join(e.args, " "), e.output)
}

func New(environment string) (*DBAAS, error) {
	targetKubeConfig := os.Getenv("HOME") + "/.percona/" + environment + "/kubeconfig"
	if len(environment) > 0 {
		if _, err := exec.LookPath(targetKubeConfig); err != nil {
			return &DBAAS{
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
	return &DBAAS{
		environment: os.Getenv("HOME") + "/.kube/" + "/config",
	}, nil
}

func (p DBAAS) runCmd(cmd string, args ...string) ([]byte, error) {
	cli := exec.Command(cmd, args...)
	cli.Env = os.Environ()
	cli.Env = append(cli.Env, "KUBECONFIG="+p.environment)

	o, err := cli.CombinedOutput()
	if err != nil {
		return nil, ErrCmdRun{cmd: cmd, args: args, output: o}
	}

	return o, nil
}

func (p DBAAS) readOperatorLogs(operatorName string) ([]byte, error) {
	return p.runCmd(execCommand, "logs", "-l", "name="+operatorName)
}

func (p DBAAS) GetObject(typ, name string) ([]byte, error) {
	return p.runCmd(execCommand, "get", typ+"/"+name, "-o", "json")
}

func (p DBAAS) apply(k8sObj string) error {
	_, err := p.runCmd("sh", "-c", "cat <<-EOF | "+execCommand+" apply -f -\n"+k8sObj+"\nEOF")
	if err != nil {
		return err
	}

	return nil
}

func (p DBAAS) IsObjExists(typ, name string) (bool, error) {
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

	out, err := p.runCmd(execCommand, "get", typ, name, "-o", "name")
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return false, errors.Wrapf(err, "get cr: %s", out)
	}

	return strings.TrimSpace(string(out)) == typ+"/"+name, nil
}

func (p DBAAS) Instances(typ string) ([]string, error) {
	out, err := p.runCmd(execCommand, "get", typ, "-o", "name")
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return nil, errors.Wrapf(err, "get objects: %s", out)
	}

	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
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
func (p DBAAS) GetPlatformType() PlatformType {
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

func (p DBAAS) checkMinikube() bool {
	output, err := p.runCmd(execCommand, "get", "storageclass", "-o", "jsonpath='{.items..provisioner}'")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "k8s.io/minikube-hostpath")
}

func (p DBAAS) checkMinishift() bool {
	output, err := p.runCmd(execCommand, "get", "pods", "master-etcd-localhost", "-n", "kube-system", "-o", "jsonpath='{.spec.volumes..path}'")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "minishift")
}

func (p DBAAS) checkOpenshift() bool {
	output, err := p.runCmd(execCommand, "api-versions")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "openshift")
}

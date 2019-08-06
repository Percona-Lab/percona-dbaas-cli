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

package gcloud

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())

	execCommand = gcloudExec
	if _, err := exec.LookPath(execCommand); err != nil {
		fmt.Println("Can't find gcloud executable. Installing it according to https://cloud.google.com/sdk/docs/downloads-interactive.")
		runEnvCmd([]string{"CLOUDSDK_CORE_DISABLE_PROMPTS=1"}, "bash", "-c", "curl https://sdk.cloud.google.com | bash")

		if _, err := exec.LookPath(execCommand); err != nil {
			panic(fmt.Sprintf("Something went wrong. Unable to find '%s' executable after installation.", gcloudExec))
		}
	}

	execCommand = k8sExecDefault
	if _, err := exec.LookPath(execCommand); err != nil {
		execCommand = k8sExecCustom
		if _, err := exec.LookPath(execCommand); err != nil {

			fmt.Println("Can't find kubectl executable. Installing it according to https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-on-linux.")
			runCmd("curl", "-LO", "https://storage.googleapis.com/kubernetes-release/release/v1.15.0/bin/linux/amd64/kubectl")
			runCmd("chmod", "+x", "./kubectl")
			runCmd("mv", "./kubectl", "/usr/local/bin/kubectl")

			if _, err := exec.LookPath(k8sExecDefault); err != nil {
				panic(fmt.Sprintf("Something went wrong. Unable to find '%s' executable after installation.", k8sExecDefault))
			}
		}
	}
	runCmd("bash", "-c", "mkdir -vp ${HOME}/.percona")
}

type PlatformType string

const (
	PlatformKubernetes PlatformType = "kubernetes"
	PlatformMinikube   PlatformType = "minikube"
	PlatformOpenshift  PlatformType = "openshift"
	PlatformMinishift  PlatformType = "minishift"
)

var execCommand string

type ErrCmdRun struct {
	cmd    string
	args   []string
	output []byte
}

func (e ErrCmdRun) Error() string {
	return fmt.Sprintf("failed to run `%s %s`, output: %s", e.cmd, strings.Join(e.args, " "), e.output)
}

func runEnvCmd(Env []string, cmd string, args ...string) ([]byte, error) {
	cli := exec.Command(cmd, args...)
	cli.Env = os.Environ()

	for _, envVar := range Env {
		cli.Env = append(cli.Env, envVar)
	}
	o, err := cli.CombinedOutput()
	if err != nil {
		return nil, ErrCmdRun{cmd: cmd, args: args, output: o}
	}

	return o, nil
}

func runCmd(cmd string, args ...string) ([]byte, error) {
	o, err := exec.Command(cmd, args...).CombinedOutput()
	if err != nil {
		return nil, ErrCmdRun{cmd: cmd, args: args, output: o}
	}

	return o, nil
}

func readOperatorLogs(operatorName string) ([]byte, error) {
	return runCmd(execCommand, "logs", "-l", "name="+operatorName)
}

func GetObject(typ, name string) ([]byte, error) {
	return runCmd(execCommand, "get", typ+"/"+name, "-o", "json")
}

func apply(k8sObj string) error {
	_, err := runCmd("sh", "-c", "cat <<-EOF | "+execCommand+" apply -f -\n"+k8sObj+"\nEOF")
	if err != nil {
		return err
	}

	return nil
}

func IsObjExists(typ, name string) (bool, error) {
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

	out, err := runCmd(execCommand, "get", typ, name, "-o", "name")
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return false, errors.Wrapf(err, "get cr: %s", out)
	}

	return strings.TrimSpace(string(out)) == typ+"/"+name, nil
}

func Instances(typ string) ([]string, error) {
	out, err := runCmd(execCommand, "get", typ, "-o", "name")
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
func GetPlatformType() PlatformType {
	if checkMinikube() {
		return PlatformMinikube
	}

	if checkMinishift() {
		return PlatformMinishift
	}

	if checkOpenshift() {
		return PlatformOpenshift
	}

	return PlatformKubernetes
}

func checkMinikube() bool {
	output, err := runCmd(execCommand, "get", "storageclass", "-o", "jsonpath='{.items..provisioner}'")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "k8s.io/minikube-hostpath")
}

func checkMinishift() bool {
	output, err := runCmd(execCommand, "get", "pods", "master-etcd-localhost", "-n", "kube-system", "-o", "jsonpath='{.spec.volumes..path}'")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "minishift")
}

func checkOpenshift() bool {
	output, err := runCmd(execCommand, "api-versions")
	if err != nil {
		return false
	}

	return strings.Contains(string(output), "openshift")
}

func isFileExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	}
	return false
}

func checkExecution(e error, tag string) (string, error) {
	if e != nil {
		return "", errors.Wrap(e, tag)
	}
	return "", errors.Wrap(nil, "empty")
}

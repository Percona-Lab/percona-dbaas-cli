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

type ErrCmdRun struct {
	cmd    string
	args   []string
	output []byte
}

func (e ErrCmdRun) Error() string {
	return fmt.Sprintf("failed to run `%s %s`, output: %s", e.cmd, strings.Join(e.args, " "), e.output)
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

func GetObject(typ, clusterName string) ([]byte, error) {
	return runCmd(execCommand, "get", typ+"/"+clusterName, "-o", "json")
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
	}

	out, err := runCmd(execCommand, "get", typ, name, "-o", "name")
	if err != nil && !strings.Contains(err.Error(), "NotFound") {
		return false, errors.Wrapf(err, "get cr: %s", out)
	}

	return strings.TrimSpace(string(out)) == typ+"/"+name, nil
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
	minikube := checkMinikube()
	if minikube {
		return PlatformMinikube
	}

	minishift := checkMinishift()
	if minishift {
		return PlatformMinishift
	}

	openshift := checkOpenshift()
	if openshift {
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

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
)

// Gcloud env structure for GKE
type Gcloud struct {
	envName     string
	project     string
	zone        string
	cluster     string
	keyfile     string
	namespace   string
	execCommand string
}

// New gcloud object
func New(name string, project string, zone string, cluster string, keyfile string, namespace string) (*Gcloud, error) {
	perconaHome := os.Getenv("HOME") + "/.percona"
	gcloudobj := &Gcloud{}

	gcloudobj.execCommand = k8sExecDefault
	if _, err := exec.LookPath(gcloudobj.execCommand); err != nil {
		gcloudobj.execCommand = k8sExecCustom
		if _, err := exec.LookPath(gcloudobj.execCommand); err != nil {

			fmt.Println("Can't find kubectl executable. Installing it according to https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-on-linux.")
			_, err := exec.Command("curl", "-L", "-o", "/usr/local/bin/kubectl", "https://storage.googleapis.com/kubernetes-release/release/v1.15.0/bin/linux/amd64/kubectl").CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("binary download has been failed. %v", err)
			}

			err = os.Chmod("/usr/local/bin/kubectl", 0755)
			if err != nil {
				return nil, fmt.Errorf("setting exec attribute failed. %v", err)
			}

			if _, err := exec.LookPath(k8sExecDefault); err != nil {
				return nil, fmt.Errorf("something went wrong. Unable to find '%s' executable after installation", k8sExecDefault)
			}
		}
	}
	gcloudobj.execCommand = gcloudExec
	if _, err := exec.LookPath(gcloudobj.execCommand); err != nil {
		fmt.Println("Can't find gcloud executable. Installing it according to https://cloud.google.com/sdk/docs/downloads-interactive.")
		runEnvCmd([]string{"CLOUDSDK_CORE_DISABLE_PROMPTS=1"}, "bash", "-c", "curl https://sdk.cloud.google.com | bash")

		if _, err := exec.LookPath(gcloudobj.execCommand); err != nil {
			return nil, fmt.Errorf("something went wrong. Unable to find '%s' executable after installation", gcloudExec)
		}
	}

	if _, err := os.Stat(perconaHome); os.IsNotExist(err) {
		err := os.Mkdir(perconaHome, 0755)
		if err != nil {
			return nil, fmt.Errorf("unable to create base .percona dir %v", err)
		}
	}
	gcloudobj.envName = name
	gcloudobj.project = project
	gcloudobj.zone = zone
	gcloudobj.cluster = cluster
	gcloudobj.keyfile = name
	gcloudobj.namespace = keyfile
	return gcloudobj, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

func isFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

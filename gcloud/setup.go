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
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

// Setup new gcloud cluster environment
func Setup(envName string, f *pflag.FlagSet) (string, error) {
	var keyFlag []byte

	kubeNamespaceFlag, err := f.GetString("kube-namespace")
	checkExecution(err, "kube-namespace")
	clusterNameFlag, err := f.GetString("gcloud-cluster")
	checkExecution(err, "gcloud-cluster")
	keyFlagEnc, err := f.GetString("gcloud-key-file")
	checkExecution(err, "gcloud-key-file")

	if len(keyFlagEnc) > 0 {
		keyFlag, err = b64.StdEncoding.DecodeString(keyFlagEnc)
		checkExecution(err, "gcloud-key-file")
	}

	zoneFlag, err := f.GetString("gcloud-zone")
	checkExecution(err, "gcloud-zone")
	projectNameFlag, err := f.GetString("gcloud-project")
	checkExecution(err, "gcloud-project")

	home := os.Getenv("HOME")
	envPath, err := filepath.Abs(home + "/.percona/" + envName)
	checkExecution(err, "envPath")

	os.Mkdir("mkdir "+envPath, 0644)
	kubeConfigEnv := fmt.Sprintf("KUBECONFIG=%s/.percona/%s/kubeconfig", home, home)

	if len(keyFlagEnc) > 0 {
		keyFilePath := envPath + "/key-file.json"
		err = ioutil.WriteFile(keyFilePath, keyFlag, 0644)
		checkExecution(err, "write file")
		_, err := runEnvCmd([]string{kubeConfigEnv}, "bash", "-c", "gcloud auth activate-service-account --key-file "+keyFilePath)
		checkExecution(err, "gcloud authentication")
	}

	_, err = runEnvCmd([]string{kubeConfigEnv}, "bash", "-c", fmt.Sprintf("gcloud container clusters get-credentials --project %s --zone %s %s", projectNameFlag, zoneFlag, clusterNameFlag))
	checkExecution(err, "getting credentials")

	_, err = runEnvCmd([]string{kubeConfigEnv}, "bash", "-c", "kubectl config set-context \"$(kubectl config current-context)\" --namespace="+kubeNamespaceFlag)
	checkExecution(err, "setting kube context")

	input, err := ioutil.ReadFile(fmt.Sprintf("%s/.kube/config", home))
	checkExecution(err, "reading kube config")
	err = ioutil.WriteFile(fmt.Sprintf("%s/kubeconfig", envPath), input, 0644)
	checkExecution(err, "writing kube config")

	status, err := SetDefaultEnv(envName, home)
	return status, err
}

// Setting default kubernetes environment
func SetDefaultEnv(envName string, homePath string) (string, error) {
	if !isFileExists(fmt.Sprintf("%s/.percona/%s", homePath, envName)) {
		return "", fmt.Errorf(fmt.Sprintf("Environment %s does not exist", envName))
	}

	if isFileExists(fmt.Sprintf("%s/.kube/config", homePath)) {
		err := os.Rename(fmt.Sprintf("%s/.kube/config", homePath), fmt.Sprintf("%s/.kube/config-old", homePath))
		checkExecution(err, "moving kube config")
	}

	err := os.Symlink(fmt.Sprintf("%s/.percona/%s/kubeconfig", homePath, envName), fmt.Sprintf("%s/.kube/config", homePath))
	checkExecution(err, "setting default kube config")
	return "", err
}

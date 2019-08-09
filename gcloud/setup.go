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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const randomStringLength = 5

// Setup new gcloud cluster environment
func (p *Gcloud) Setup() error {
	homePath := os.Getenv("HOME")
	envPath, err := filepath.Abs(homePath + "/.percona/" + p.envName)
	if err != nil {
		return errors.Wrap(err, "environment")
	}
	fileExists, err := isFileExists(envPath)
	if err != nil {
		return errors.Wrap(err, "access file")
	}
	if !fileExists {
		err = os.Mkdir(envPath, 0755)
		if err != nil {
			return errors.Wrap(err, "create directory")
		}
	}
	kubeConfigEnv := fmt.Sprintf("KUBECONFIG=%s/.percona/%s/kubeconfig", homePath, p.envName)

	if len(p.keyfile) > 0 {
		jsonKey, err := base64.StdEncoding.DecodeString(p.keyfile)
		if err != nil {
			return errors.Wrap(err, "key-file decode")
		}
		keyFilePath := envPath + "/key-file.json"
		err = ioutil.WriteFile(keyFilePath, jsonKey, 0400)
		if err != nil {
			return errors.Wrap(err, "write file")
		}
		_, err = runEnvCmd([]string{kubeConfigEnv}, "bash", "-c", "gcloud auth activate-service-account --key-file "+keyFilePath)
		if err != nil {
			return errors.Wrap(err, "gcloud authentication")
		}
	}
	_, err = runEnvCmd([]string{kubeConfigEnv}, "bash", "-c", fmt.Sprintf("gcloud container clusters get-credentials --project %s --zone %s %s", p.project, p.zone, p.cluster))
	if err != nil {
		return errors.Wrap(err, "getting credentials")
	}
	_, err = runEnvCmd([]string{kubeConfigEnv}, "bash", "-c", "kubectl config set-context \"$(kubectl config current-context)\" --namespace="+p.namespace)
	if err != nil {
		return errors.Wrap(err, "setting kube context")
	}
	return SetDefaultEnv(p.envName, homePath)
}

// SetDefaultEnv sets default kubernetes environment
func SetDefaultEnv(envName string, homePath string) error {
	fileExists, err := isFileExists(fmt.Sprintf("%s/.percona/%s", homePath, envName))
	if err != nil {
		return errors.Wrap(err, "access file")
	}
	if !fileExists {
		return fmt.Errorf("environment %s does not exist", envName)
	}

	fileExists, err = isFileExists(fmt.Sprintf("%s/.kube/config", homePath))
	if err != nil {
		return errors.Wrap(err, "access file")
	}
	if fileExists {
		err := os.Rename(fmt.Sprintf("%s/.kube/config", homePath), fmt.Sprintf("%s/.kube/config-sdc33s-old", homePath))
		if err != nil {
			return errors.Wrap(err, "moving kube config")
		}
	}
	err = os.Symlink(fmt.Sprintf("%s/.percona/%s/kubeconfig", homePath, envName), fmt.Sprintf("%s/.kube/config", homePath))
	if err != nil {
		return errors.Wrap(err, "setting default kube config")
	}
	return nil
}

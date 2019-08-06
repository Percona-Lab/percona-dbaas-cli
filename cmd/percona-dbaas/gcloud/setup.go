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
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/Percona-Lab/percona-dbaas-cli/gcloud"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}

func checkExecution(e error) {
	if e != nil {
		fmt.Printf("\n[error] %s\n", e)
		return
	}
}

// setupCmd represents the list command
var setupCmd = &cobra.Command{
	Use:   "setup <environment-name>",
	Short: "Setup your GKE cluster credentials",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify environment name")
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		args = parseArgs(args)

		envName := args[0]
		keyFlag, err := cmd.Flags().GetString("gcloud-key-file")
		checkExecution(err)
		if keyFlag != "" {
			if isJSON(keyFlag) {
				cmd.Flags().Set("gcloud-key-file", b64.StdEncoding.EncodeToString([]byte(keyFlag)))
			} else {
				dat, err := ioutil.ReadFile(keyFlag)
				checkExecution(err)
				cmd.Flags().Set("gcloud-key-file", b64.StdEncoding.EncodeToString([]byte(dat)))
			}

		}
		_, err = gcloud.Setup(envName, cmd.Flags())
		checkExecution(err)
	},
}

func init() {
	setupCmd.Flags().String("gcloud-project", "", "Google Cloud organization unit")
	setupCmd.Flags().String("gcloud-zone", "", "Compute zone (us-central1-a, europe-west3-b ...)")
	setupCmd.Flags().String("gcloud-cluster", "", "Google Kubernetes Engine cluster name")
	setupCmd.Flags().String("gcloud-key-file", "", "Google Cloud credentials file path or the contents.")
	setupCmd.Flags().String("kube-namespace", "default", "Default kubernetes namespace")

	setupCmd.MarkFlagRequired("gcloud-project")
	setupCmd.MarkFlagRequired("gcloud-zone")
	setupCmd.MarkFlagRequired("gcloud-cluster")

	GCLOUDCmd.AddCommand(setupCmd)
}

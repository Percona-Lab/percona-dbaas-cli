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
		envName := args[0]
		if len(*keyfile) > 0 {
			if isJSON(*keyfile) {
				*keyfile = base64.StdEncoding.EncodeToString([]byte(*keyfile))
			} else {
				dat, err := ioutil.ReadFile(*keyfile)
				if err != nil {
					fmt.Printf("\n[error] %s\n", err)
					return
				}
				*keyfile = base64.StdEncoding.EncodeToString([]byte(dat))
			}

		}

		cloudEnv, err := gcloud.New(envName, *project, *zone, *cluster, *keyfile, *namespace)
		if err != nil {
			fmt.Printf("\n[error] %s\n", err)
			return
		}
		err = cloudEnv.Setup()
		if err != nil {
			fmt.Printf("\n[error] %s\n", err)
			return
		}
	},
}

var project *string
var zone *string
var cluster *string
var keyfile *string
var namespace *string

func init() {
	project = setupCmd.Flags().String("gcloud-project", "", "Google Cloud organization unit")
	zone = setupCmd.Flags().String("gcloud-zone", "", "Compute zone (us-central1-a, europe-west3-b ...)")
	cluster = setupCmd.Flags().String("gcloud-cluster", "", "Google Kubernetes Engine cluster name")
	keyfile = setupCmd.Flags().String("gcloud-key-file", "", "Google Cloud credentials file path or the contents.")
	namespace = setupCmd.Flags().String("kube-namespace", "default", "Default kubernetes namespace")

	setupCmd.MarkFlagRequired("gcloud-project")
	setupCmd.MarkFlagRequired("gcloud-zone")
	setupCmd.MarkFlagRequired("gcloud-cluster")

	GCLOUDCmd.AddCommand(setupCmd)
}

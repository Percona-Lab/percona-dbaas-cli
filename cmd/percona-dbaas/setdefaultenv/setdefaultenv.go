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

package setdefaultenv

import (
	"errors"
	"fmt"
	"os"

	"github.com/Percona-Lab/percona-dbaas-cli/gcloud"
	"github.com/spf13/cobra"
)

// SetDefaultEnvCmd represents the gcloud command
var SetDefaultEnvCmd = &cobra.Command{
	Use:   "set-default-environment",
	Short: "Set your kubernetes/openshift environment as default",

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify environment name")
		}

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {

		envName := args[0]
		err := gcloud.SetDefaultEnv(envName, os.Getenv("HOME"))
		if err != nil {
			fmt.Printf("\n[error] %s\n", err)
			return
		}
	},
}

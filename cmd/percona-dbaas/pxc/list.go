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

package pxc

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "describe-db <db-cluster-name>",
	Short: "Show either specific MySQL cluster or all clusters on current Kubernetes environment",

	Run: func(cmd *cobra.Command, args []string) {
		app, err := pxc.New(args[0], defaultVersion, "", *envLst)
		if err != nil {
			pxc.PrintError(*listAnswerOutput, "new operator", err)
			return
		}
		if len(args) > 0 {
			info, err := app.Describe()
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
				return
			}
			fmt.Print(info)
			return
		}
		list, err := app.List()
		if err != nil {
			pxc.PrintError(*listAnswerOutput, "pxc list", err)
			return
		}

		fmt.Print(list)
	},
}

var envLst *string
var listAnswerOutput *string

func init() {
	envLst = listCmd.Flags().String("environment", "", "Target kubernetes cluster")
	listAnswerOutput = listCmd.Flags().String("output", "", "Output format")

	PXCCmd.AddCommand(listCmd)
}

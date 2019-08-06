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

package psmdb

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "describe-db",
	Short: "Show either specific MongoDB cluster or all clusters on current Kubernetes environment",

	Run: func(cmd *cobra.Command, args []string) {
		list, err := dbaas.List("psmdb")
		if err != nil {
			fmt.Printf("\n[error] %s\n", err)
			return
		}

		fmt.Print(list)
	},
}

func init() {
	PSMDBCmd.AddCommand(listCmd)
}

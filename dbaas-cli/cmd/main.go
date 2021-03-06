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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/cmd/mongo"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/cmd/mysql"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "percona-dbaas",
	Short: "The simplest DBaaS tool in the world",
	Long: `    Hello, it is the simplest DBaaS tool in the world,
	please use commands below to manage your DBaaS.`,
}

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "text", `Answers format. Can be "json" or "text".`)
	rootCmd.AddCommand(mysql.PXCCmd)
	rootCmd.AddCommand(mongo.MongoCmd)
	rootCmd.PersistentFlags().Bool("no-wait", false, "Dont wait while command is done")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

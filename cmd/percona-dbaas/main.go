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
	"path"

	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/cmd/percona-dbaas/gcloud"
	"github.com/Percona-Lab/percona-dbaas-cli/cmd/percona-dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/cmd/percona-dbaas/pxc"
	broker "github.com/Percona-Lab/percona-dbaas-cli/cmd/percona-dbaas/service-broker"
	"github.com/Percona-Lab/percona-dbaas-cli/cmd/percona-dbaas/setdefaultenv"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "percona-dbaas",
	Short: "The simplest DBaaS tool in the world",
	Long: `    Hello, it is the simplest DBaaS tool in the world,
	please use commands below to manage your DBaaS.`,
}

func init() {
	rootCmd.PersistentFlags().Bool("demo", false, "demo mode (no spinners)")
	rootCmd.PersistentFlags().MarkHidden("demo")
	rootCmd.PersistentFlags().String("output", "", `Answers format. Can be only "json"`)

	rootCmd.AddCommand(pxc.PXCCmd)
	rootCmd.AddCommand(psmdb.PSMDBCmd)
	rootCmd.AddCommand(broker.PxcBrokerCmd)
	rootCmd.AddCommand(gcloud.GCLOUDCmd)
	rootCmd.AddCommand(setdefaultenv.SetDefaultEnvCmd)
}

func main() {
	rewriteKubectlArgs("pxc")
	rewriteKubectlArgs("psmdb")
	rewriteKubectlArgs("gcloud")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func rewriteKubectlArgs(db string) {
	if path.Base(os.Args[0]) == "kubectl-"+db {
		os.Args = append(os.Args[:1], append([]string{db}, os.Args[1:]...)...)
	}
}

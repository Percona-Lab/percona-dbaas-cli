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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "percona-dbaas",
	Short: "The simplest DBaaS tool in the world",
	Long: `    Hello, it is the simplest DBaaS tool in the world,
    please use commands below to manage your DBaaS.
 
    # percona-dbaas pxc create <pxc-cluster-name>
        for create MySQL cluster on current Kubernetes cluster
    # percona-dbaas pxc list
        list all MySQL clusters on current Kubernetes cluster
    # percona-dbaas pxc delete <pxc-cluster-name>
        delete MySQL cluster
    # percona-dbaas pxc create-backup <pxc-cluster-name>
        create on-demand backup of MySQL cluster`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

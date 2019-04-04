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
	"time"

	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const (
	defaultVersion = "0.3.0"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create MySQL cluster on current Kubernetes cluster",

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Printf("[Error] you have to define pxc-cluster-name")
			return
		}
		app, err := pxc.New(args[0], defaultVersion)

		if err != nil {
			fmt.Printf("[Error] create pxc: %v", err)
			return
		}

		err = app.SetConfig(cmd.Flags())
		if err != nil {
			fmt.Printf("[Error] set configuration: %v", err)
			return
		}

		fmt.Print("\n\nStarting")

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		go dbaas.Create(app, created, msg, cerr)
		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for range tckr.C {
			select {
			case okmsg := <-created:
				fmt.Printf("[done]\n\n%s\n", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					fmt.Printf("\n[error] %s\n", omsg)
				}

			case err := <-cerr:
				fmt.Fprintf(os.Stderr, "[ERROR] create pxc: %v\n", err)
				return
			default:
				fmt.Print(".")
			}
		}
	},
}

func init() {
	createCmd.Flags().String("storage", "6G", "PXC node volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)")
	createCmd.Flags().Int("pxc-instances", 3, "Number of PXC nodes in cluster")
	createCmd.Flags().String("pxc-request-cpu", "600m", "PXC node requests for CPU, in cores. (500m = .5 cores)")
	createCmd.Flags().String("pxc-request-mem", "1G", "PXC node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")

	createCmd.Flags().String("proxy-storage", "2G", "ProxySQL node volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)")
	createCmd.Flags().Int("proxy-instances", 1, "Number of ProxySQL nodes in cluster")
	createCmd.Flags().String("proxy-request-cpu", "600m", "ProxySQL node requests for CPU, in cores. (500m = .5 cores)")
	createCmd.Flags().String("proxy-request-mem", "1G", "ProxySQL node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")

	pxcCmd.AddCommand(createCmd)
}

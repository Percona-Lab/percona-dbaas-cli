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

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit <pxc-cluster-name>",
	Short: "Change MySQL cluster on current Kubernetes cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		app, err := pxc.New(args[0], defaultVersion)

		if err != nil {
			fmt.Printf("[Error] create pxc: %v", err)
			return
		}

		setupmsg, err := app.Setup(cmd.Flags())
		if err != nil {
			fmt.Printf("[Error] set configuration: %v", err)
			return
		}

		fmt.Print("Looking for cluster")
		checkDone := make(chan struct{})
		var ext bool
		go func() {
			ext, err = dbaas.IsCRexists("pxc", app.ClusterName())
			checkDone <- struct{}{}
		}()

		dtckr := time.NewTicker(1 * time.Second)
		defer dtckr.Stop()
	CHECKLOOP:
		for range dtckr.C {
			select {
			case <-checkDone:
				fmt.Println("[done]")
				break CHECKLOOP
			default:
				fmt.Print(".")
			}
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] check if cluster exists: %v\n", err)
			return
		}

		if !ext {
			fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"", "pxc", app.ClusterName())

			list, err := dbaas.List("pxc")
			if err != nil {
				return
			}

			fmt.Println("\nAvaliable clusters:")
			fmt.Print(list)

			return
		}

		fmt.Printf("%s\nStarting", setupmsg)

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		go dbaas.Edit("pxc", app, created, msg, cerr)
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
				fmt.Fprintf(os.Stderr, "\n[ERROR] create pxc: %v\n", err)
				return
			default:
				fmt.Print(".")
			}
		}
	},
}

func init() {
	// editCmd.Flags().Int("pxc-instances", 3, "Number of PXC nodes in cluster")
	// editCmd.Flags().String("pxc-request-cpu", "600m", "PXC node requests for CPU, in cores. (500m = .5 cores)")
	// editCmd.Flags().String("pxc-request-mem", "1G", "PXC node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	// editCmd.Flags().String("pxc-request-cpu", "600m", "PXC node requests for CPU, in cores. (500m = .5 cores)")
	// editCmd.Flags().String("pxc-request-mem", "1G", "PXC node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")

	// editCmd.Flags().Int("proxy-instances", 1, "Number of ProxySQL nodes in cluster")
	// editCmd.Flags().String("proxy-request-cpu", "600m", "ProxySQL node requests for CPU, in cores. (500m = .5 cores)")
	// editCmd.Flags().String("proxy-request-mem", "1G", "ProxySQL node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")

	// pxcCmd.AddCommand(editCmd)
}

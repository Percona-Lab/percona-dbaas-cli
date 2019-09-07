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
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "modify-db <pxc-cluster-name>",
	Short: "Modify MySQL cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		app, err := pxc.New(name, defaultVersion, "", *envEdt)
		if err != nil {
			pxc.PrintError(*editAnswerOutput, "new operator", err)
			return
		}

		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		demo, err := cmd.Flags().GetBool("demo")
		if demo && err == nil {
			sp.UpdateCharSet([]string{""})
		}
		sp.Prefix = "Looking for the cluster..."
		sp.FinalMSG = ""
		sp.Start()
		defer sp.Stop()

		ext, err := app.Cmd.IsObjExists("pxc", name)
		if err != nil {
			pxc.PrintError(*editAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			pxc.PrintError(*editAnswerOutput, "unable to find cluster pxc/"+name, nil)
			list, err := app.Cmd.List("pxc")
			if err != nil {
				pxc.PrintError(*editAnswerOutput, "pxc clusters list", err)
				return
			}
			fmt.Fprint(os.Stderr, "Avaliable clusters:")
			fmt.Print(list)
			return
		}

		config, err := pxc.ParseEditFlagsToConfig(cmd.Flags())
		if err != nil {
			pxc.PrintError(*editAnswerOutput, "parse flags to config", err)
			return
		}
		app.ClusterConfig = config

		created := make(chan pxc.ClusterData)
		msg := make(chan pxc.ClusterData)
		cerr := make(chan error)
		go app.Edit(nil, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Applying changes..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := app.Cmd.ListName("pxc", name)
				finalMsg, err := SprintResponse(*editAnswerOutput, okmsg)
				if err != nil {
					pxc.PrintError(*editAnswerOutput, "sprint response", err)
				}
				sp.FinalMSG = fmt.Sprintln("Applying changes...[done]\n\n", finalMsg)
				return
			case omsg := <-msg:
				sp.Stop()
				pxc.PrintError(*editAnswerOutput, "operator log error: "+omsg.Message, nil)
				sp.Start()
			case err := <-cerr:
				pxc.PrintError(*editAnswerOutput, "edit pxc", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envEdt *string
var editAnswerOutput *string

func init() {
	editCmd.Flags().Int32("pxc-instances", 0, "Number of PXC nodes in cluster")
	editCmd.Flags().Int32("proxy-instances", -1, "Number of ProxySQL nodes in cluster")
	envEdt = editCmd.Flags().String("environment", "", "Target kubernetes cluster")
	editAnswerOutput = editCmd.Flags().String("output", "", "Output format")

	PXCCmd.AddCommand(editCmd)
}

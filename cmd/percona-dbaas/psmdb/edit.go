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
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "modify-db <psmdb-cluster-name>",
	Short: "Modify MongoDB cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		args = parseArgs(args)

		clusterName := args[0]

		rsName := ""
		if len(args) >= 2 {
			rsName = args[1]
		}

		app, err := psmdb.New(clusterName, rsName, defaultVersion, false, "", *envEdt)
		if err != nil {
			psmdb.PrintError(*editAnswerOutput, "new psmdb app", err)
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

		ext, err := app.Cmd.IsObjExists("psmdb", clusterName)
		if err != nil {
			psmdb.PrintError(*editAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			psmdb.PrintError(*editAnswerOutput, "unable to find cluster psmdb/"+clusterName, nil)
			list, err := app.Cmd.List("psmdb")
			if err != nil {
				psmdb.PrintError(*editAnswerOutput, "psmdb clusters", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)
		config, err := psmdb.ParseEditFlagsToConfig(cmd.Flags())
		if err != nil {
			psmdb.PrintError(*editAnswerOutput, "parsing flags", err)
			return
		}
		app.ClusterConfig = config
		go app.Edit(nil, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Applying changes..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := app.Cmd.ListName("psmdb", clusterName)
				sp.FinalMSG = fmt.Sprintf("Applying changes...[done]\n\n%s", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					psmdb.PrintError(*editAnswerOutput, "operator log error", fmt.Errorf(omsg.String()))
					sp.Start()
				}
			case err := <-cerr:
				psmdb.PrintError(*editAnswerOutput, "edit psmdb", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envEdt *string
var editAnswerOutput *string

func init() {
	editCmd.Flags().Int32("replset-size", 3, "Number of nodes in replset")
	envEdt = editCmd.Flags().String("environment", "", "Target kubernetes cluster")
	editAnswerOutput = editCmd.Flags().String("output", "", "Output format")

	PSMDBCmd.AddCommand(editCmd)
}

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
	"os"
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

		dbservice, err := dbaas.New(*envEdt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			return
		}
		rsName := ""
		if len(args) >= 2 {
			rsName = args[1]
		}

		app := psmdb.New(clusterName, rsName, defaultVersion, *editAnswerInJSON)

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

		ext, err := dbservice.IsObjExists("psmdb", clusterName)

		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] check if cluster exists: %v\n", err)
			return
		}

		if !ext {
			sp.Stop()
			fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "psmdb", clusterName)
			list, err := dbservice.List("psmdb")
			if err != nil {
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
			fmt.Println("Parsing flags", err)
		}
		app.ClusterConfig = config
		go dbservice.Edit("psmdb", app, nil, created, msg, cerr)
		sp.Prefix = "Applying changes..."

		for {
			select {
			case <-created:
				okmsg, _ := dbservice.ListName("psmdb", clusterName)
				sp.FinalMSG = fmt.Sprintf("Applying changes...[done]\n\n%s", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					fmt.Printf("[operator log error] %s\n", omsg)
					sp.Start()
				}
			case err := <-cerr:
				fmt.Fprintf(os.Stderr, "\n[ERROR] edit psmdb: %v\n", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envEdt *string
var editAnswerInJSON *bool

func init() {
	editCmd.Flags().Int32("replset-size", 3, "Number of nodes in replset")
	envEdt = editCmd.Flags().String("environment", "", "Target kubernetes cluster")

	editAnswerInJSON = editCmd.Flags().Bool("json", false, "Answers in JSON format")

	PSMDBCmd.AddCommand(editCmd)
}

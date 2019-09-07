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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// upgradeOperatorCmd represents the edit command
var upgradeOperatorCmd = &cobra.Command{
	Use:   "upgrade-operator <psmdb-cluster-name> <to-version>",
	Short: "Upgrade PSMDB operator",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		app, err := psmdb.New(name, "doesnotMatter", defaultVersion, "", *envUpgrdOprtr)
		if err != nil {
			psmdb.PrintError(*upgradeOperatorAnswerOutput, "new psmdb operator", err)
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

		ext, err := app.Cmd.IsObjExists("psmdb", name)

		if err != nil {
			psmdb.PrintError(*upgradeOperatorAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "psmdb", name)
			list, err := app.List()
			if err != nil {
				psmdb.PrintError(*upgradeOperatorAnswerOutput, "db service list", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		created := make(chan string)
		cerr := make(chan error)

		if *oprtrImage != "" {
			num, err := app.Cmd.Instances("psmdb")
			if err != nil {
				psmdb.PrintError(*upgradeOperatorAnswerOutput, "unable to get psmdb instances", err)
				return
			}
			if len(num) > 1 {
				sp.Stop()
				var yn string
				fmt.Printf("\nFound more than one psmdb cluster: %v.\nOperator upgrade may affect other clusters.\nContinue? [y/N] ", num)
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					yn = strings.TrimSpace(scanner.Text())
					break
				}
				if yn != "y" && yn != "Y" {
					return
				}
				sp.Start()
			}
		}

		go app.UpgradeOperator(*oprtrImage, created, cerr)
		sp.Lock()
		sp.Prefix = "Upgrading cluster operator..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := app.Cmd.ListName("psmdb", name)
				finalMsg, err := SprintResponse(*upgradeOperatorAnswerOutput, okmsg)
				if err != nil {
					pxc.PrintError(*upgradeOperatorAnswerOutput, "sprint response", err)
				}
				sp.FinalMSG = fmt.Sprintln("Upgrading cluster operator...[done]\n\n", finalMsg)
				return
			case err := <-cerr:
				psmdb.PrintError(*upgradeOperatorAnswerOutput, "upgrade psmdb operator", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envUpgrdOprtr *string
var oprtrImage *string
var upgradeOperatorAnswerOutput *string

func init() {
	oprtrImage = upgradeOperatorCmd.Flags().String("operator-image", "", "Custom image to upgrade operator to")
	envUpgrdOprtr = upgradeOperatorCmd.Flags().String("environment", "", "Target kubernetes cluster")
	upgradeOperatorAnswerOutput = upgradeOperatorCmd.Flags().String("output", "", "Output format")

	PSMDBCmd.AddCommand(upgradeOperatorCmd)
}

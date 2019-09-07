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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// upgradeOperatorCmd represents the edit command
var upgradeOperatorCmd = &cobra.Command{
	Use:   "upgrade-operator <pxc-cluster-name> <to-version>",
	Short: "Upgrade PXC operator",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		app, err := pxc.New(name, defaultVersion, "", *envUpgrdOprtr)
		if err != nil {
			pxc.PrintError(*upgradeOperatorAnswerOutput, "new operator", err)
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
			pxc.PrintError(*upgradeOperatorAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			pxc.PrintError(*upgradeOperatorAnswerOutput, "Unable to find cluster pxc/"+name, nil)
			list, err := app.Cmd.List("pxc")
			if err != nil {
				pxc.PrintError(*upgradeOperatorAnswerOutput, "list pxc", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		created := make(chan pxc.UpgradeResponse)
		cerr := make(chan error)

		if *oprtrImage != "" {
			num, err := app.Cmd.Instances("pxc")
			if err != nil {
				pxc.PrintError(*upgradeOperatorAnswerOutput, "unable to get pxc instances", err)
			}
			if len(num) > 1 {
				sp.Stop()
				var yn string
				fmt.Printf("\nFound more than one pxc cluster: %s.\nOperator upgrade may affect other clusters.\nContinue? [y/N] ", num)
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
				okmsg, _ := app.Cmd.ListName("pxc", name)
				finalMsg, err := SprintResponse(*upgradeOperatorAnswerOutput, okmsg)
				if err != nil {
					pxc.PrintError(*upgradeOperatorAnswerOutput, "sprint response", err)
				}
				sp.FinalMSG = fmt.Sprintln("Upgrading cluster operator...[done]\n\n", finalMsg)
				return
			case err := <-cerr:
				pxc.PrintError(*upgradeOperatorAnswerOutput, "upgrade pxc operator", err)
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

	PXCCmd.AddCommand(upgradeOperatorCmd)
}

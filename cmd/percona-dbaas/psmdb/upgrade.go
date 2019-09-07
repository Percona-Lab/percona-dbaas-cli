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

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// upgradeCmd represents the edit command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade-db <psmdb-cluster-name> <to-version>",
	Short: "Upgrade Percona Server for MongoDB",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		app, err := psmdb.New(name, "doesnotMatter", defaultVersion, "", *envUpgrd)
		if err != nil {
			psmdb.PrintError(*upgradeAnswerOutput, "new psmdb operator", err)
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
			psmdb.PrintError(*upgradeAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "psmdb", name)
			list, err := app.List()
			if err != nil {
				psmdb.PrintError(*upgradeAnswerOutput, "psmdb cluster list", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		created := make(chan psmdb.ClusterData)
		msg := make(chan psmdb.ClusterData)
		cerr := make(chan error)

		oparg := ""
		if len(args) > 1 {
			oparg = args[1]
		}
		appsImg, err := app.Images(oparg, cmd.Flags())
		if err != nil {
			psmdb.PrintError(*upgradeAnswerOutput, "setup images for upgrade", err)
			return
		}

		go app.Upgrade(appsImg, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Upgrading cluster..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := app.Cmd.ListName("psmdb", name)
				finalMsg, err := SprintResponse(*upgradeAnswerOutput, okmsg)
				if err != nil {
					pxc.PrintError(*upgradeAnswerOutput, "sprint response", err)
				}
				sp.FinalMSG = fmt.Sprintln("Upgrading cluster...[done]\n\n", finalMsg)
				return
			case omsg := <-msg:
				sp.Stop()
				psmdb.PrintError(*upgradeAnswerOutput, "operator log error", fmt.Errorf(omsg.Message))
				sp.Start()
			case err := <-cerr:
				psmdb.PrintError(*upgradeAnswerOutput, "upgrade psmdb", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envUpgrd *string
var upgradeAnswerOutput *string

func init() {
	upgradeCmd.Flags().String("database-image", "", "Custom image to upgrade psmdb to")
	upgradeCmd.Flags().String("backup-image", "", "Custom image to upgrade backup to")
	envUpgrd = upgradeCmd.Flags().String("environment", "", "Target kubernetes cluster")

	upgradeAnswerOutput = upgradeCmd.Flags().String("output", "", "Output format")

	PSMDBCmd.AddCommand(upgradeCmd)
}

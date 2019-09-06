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
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// upgradeCmd represents the edit command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade-db <pxc-cluster-name> <to-version>",
	Short: "Upgrade MySQL cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		app, err := pxc.New(name, defaultVersion, false, "", *envUpgrd)
		if err != nil {
			pxc.PrintError(*upgradeAnswerOutput, "new operator", err)
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
			pxc.PrintError(*upgradeAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			pxc.PrintError(*upgradeAnswerOutput, "unable to find cluster pxc/"+name, nil)
			list, err := app.Cmd.List("pxc")
			if err != nil {
				pxc.PrintError(*upgradeAnswerOutput, "list pxc clusters", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		oparg := ""
		if len(args) > 1 {
			oparg = args[1]
		}
		appsImg, err := app.Images(oparg, cmd.Flags())
		if err != nil {
			pxc.PrintError(*upgradeAnswerOutput, "setup images for upgrade", err)
			return
		}

		go app.Upgrade(appsImg, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Upgrading cluster..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := app.Cmd.ListName("pxc", name)
				sp.FinalMSG = fmt.Sprintf("Upgrading cluster...[done]\n\n%s", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					pxc.PrintError(*upgradeAnswerOutput, "operator log error", nil)
					sp.Start()
				}
			case err := <-cerr:
				pxc.PrintError(*upgradeAnswerOutput, "upgrade pxc", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envUpgrd *string
var upgradeAnswerOutput *string

func init() {
	upgradeCmd.Flags().String("database-image", "", "Custom image to upgrade pxc to")
	upgradeCmd.Flags().String("proxysql-image", "", "Custom image to upgrade proxySQL to")
	upgradeCmd.Flags().String("backup-image", "", "Custom image to upgrade backup to")
	envUpgrd = upgradeCmd.Flags().String("environment", "", "Target kubernetes cluster")

	upgradeAnswerOutput = upgradeCmd.Flags().String("output", "", "Output format")

	PXCCmd.AddCommand(upgradeCmd)
}

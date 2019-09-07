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

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// restoreCmd represents the list command
var restoreCmd = &cobra.Command{
	Use:   "restore-db <pxc-cluster-name> <pxc-backup-name>",
	Short: "Restore MySQL backup",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name and pxc-backup-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		args = parseArgs(args)

		name := args[0]
		if len(args) < 2 || args[1] == "" {
			pxc.PrintError(*backupRestoreAnswerOutput, "you have to specify pxc-cluster-name and pxc-backup-name", nil)
			return
		}
		bcpName := args[1]

		bcp, err := pxc.NewRestore(name, *envBckpRstr)
		if err != nil {
			pxc.PrintError(*backupRestoreAnswerOutput, "creating restore object", err)
			return
		}
		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		sp.Prefix = "Looking for the cluster..."
		sp.FinalMSG = ""
		sp.Start()
		defer sp.Stop()

		ext, err := bcp.Cmd.IsObjExists("pxc", name)
		if err != nil {
			pxc.PrintError(*backupRestoreAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			pxc.PrintError(*backupRestoreAnswerOutput, "unable to find cluster pxc/"+bcpName, nil)
			list, err := bcp.Cmd.List("pxc")
			if err != nil {
				pxc.PrintError(*backupRestoreAnswerOutput, "list pxc", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		sp.Prefix = "Looking for the backup..."
		ext, err = bcp.Cmd.IsObjExists("pxc-backup", bcpName)
		if err != nil {
			pxc.PrintError(*backupRestoreAnswerOutput, "check if backup exists", err)
			return
		}

		if !ext {
			sp.Stop()
			pxc.PrintError(*backupRestoreAnswerOutput, "unable to find backup pxc-backup/"+bcpName, nil)
			list, err := bcp.Cmd.List("pxc-backup")
			if err != nil {
				pxc.PrintError(*backupRestoreAnswerOutput, "list pxc-backup", err)
				return
			}
			fmt.Println("Avaliable backups:")
			fmt.Print(list)
			return
		}
		sp.Lock()
		sp.Prefix = "Restoring backup..."
		sp.Unlock()

		bcp.Setup(bcpName)

		ok := make(chan pxc.RestoreResponse)
		msg := make(chan pxc.RestoreResponse)
		cerr := make(chan error)

		go bcp.Create(ok, msg, cerr)
		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for {
			select {
			case okmsg := <-ok:
				finalMsg, err := SprintResponse(*backupRestoreAnswerOutput, okmsg)
				if err != nil {
					pxc.PrintError(*backupRestoreAnswerOutput, "sprint response", err)
				}
				sp.FinalMSG = fmt.Sprintln("Restoring backup...[done]\n\n", finalMsg)
				return
			case omsg := <-msg:
				sp.Stop()
				pxc.PrintError(*backupRestoreAnswerOutput, "loperator log error: "+omsg.Message, nil)
				sp.Start()
			case err := <-cerr:
				pxc.PrintError(*backupRestoreAnswerOutput, "restore backup", err)
				return
			}
		}
	},
}

var envBckpRstr *string
var backupRestoreAnswerOutput *string

func init() {
	envBckpRstr = restoreCmd.Flags().String("environment", "", "Target kubernetes cluster")
	backupRestoreAnswerOutput = restoreCmd.Flags().String("output", "s", "Output format")

	PXCCmd.AddCommand(restoreCmd)
}

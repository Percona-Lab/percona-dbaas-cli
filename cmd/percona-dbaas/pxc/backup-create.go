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

// bcpCmd represents the list command
var bcpCmd = &cobra.Command{
	Use:   "create-backup <pxc-cluster-name>",
	Short: "Create MySQL backup",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		bcp, err := pxc.NewBackup(name, *envBckpCrt)
		if err != nil {
			pxc.PrintError(*backupCreateAnswerOutput, "creating backup object", err)
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
			pxc.PrintError(*backupCreateAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			pxc.PrintError(*backupCreateAnswerOutput, "Unable to find cluster pxc/"+name, nil)
			list, err := bcp.Cmd.List("pxc")
			if err != nil {
				pxc.PrintError(*backupCreateAnswerOutput, "list pxc clusters", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}
		sp.Lock()
		sp.Prefix = "Creating backup..."
		sp.Unlock()

		bcp.Setup(dbaas.DefaultBcpStorageName)

		ok := make(chan pxc.BackupResponse)
		msg := make(chan pxc.BackupResponse)
		cerr := make(chan error)

		go bcp.Create(ok, msg, cerr)
		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for {
			select {
			case okmsg := <-ok:
				finalMsg, err := SprintResponse(*addStorageAnswerOutput, okmsg)
				if err != nil {
					pxc.PrintError(*addStorageAnswerOutput, "sprint response", err)
				}
				sp.FinalMSG = fmt.Sprintln("Creating backup...[done]\n\n", finalMsg)
				return
			case omsg := <-msg:
				sp.Stop()
				pxc.PrintError(*backupCreateAnswerOutput, "operator log error: "+omsg.Message, nil)
				sp.Start()
			case err := <-cerr:
				pxc.PrintError(*backupCreateAnswerOutput, "create backup", err)
				return
			}
		}
	},
}

var envBckpCrt *string
var backupCreateAnswerOutput *string

func init() {
	envBckpCrt = bcpCmd.Flags().String("environment", "", "Target kubernetes cluster")
	backupCreateAnswerOutput = bcpCmd.Flags().String("output", "s", "Output format")

	PXCCmd.AddCommand(bcpCmd)
}

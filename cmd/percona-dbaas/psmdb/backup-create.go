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
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// bcpCmd represents the list command
var bcpCmd = &cobra.Command{
	Use:   "create-backup <psmdb-cluster-name>",
	Short: "Create MongoDB backup",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		bcp, err := psmdb.NewBackup(name, *envBckpCrt)
		if err != nil {
			psmdb.PrintError(*backupCreateAnswerOutput, "creating backup object", err)
			return
		}
		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		sp.Prefix = "Looking for the cluster..."
		sp.FinalMSG = ""
		sp.Start()
		defer sp.Stop()

		ext, err := bcp.Cmd.IsObjExists("psmdb", name)
		if err != nil {
			psmdb.PrintError(*backupCreateAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			psmdb.PrintError(*backupCreateAnswerOutput, "unable to find cluster psmdb/"+name, nil)
			fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "psmdb", name)
			list, err := bcp.Cmd.List("psmdb")
			if err != nil {
				psmdb.PrintError(*backupCreateAnswerOutput, "psmdb clusters list", err)
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

		ok := make(chan psmdb.BackupResponse)
		msg := make(chan psmdb.BackupResponse)
		cerr := make(chan error)

		go bcp.Create(ok, msg, cerr)
		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for {
			select {
			case okmsg := <-ok:
				finalMsg, err := SprintResponse(*backupCreateAnswerOutput, okmsg)
				if err != nil {
					pxc.PrintError(*backupCreateAnswerOutput, "sprint response", err)
				}
				sp.FinalMSG = fmt.Sprintf("Creating backup...[done]\n%s\n", finalMsg)
				return
			case omsg := <-msg:
				sp.Stop()
				psmdb.PrintError(*backupCreateAnswerOutput, "operator log error", fmt.Errorf(omsg.Message))
				sp.Start()
			case err := <-cerr:
				psmdb.PrintError(*backupCreateAnswerOutput, "psmdb clusters list", err)
				return
			}
		}
	},
}

var envBckpCrt *string
var backupCreateAnswerOutput *string

func init() {
	envBckpCrt = bcpCmd.Flags().String("environment", "", "Target kubernetes cluster")
	backupCreateAnswerOutput = bcpCmd.Flags().String("output", "", "Output format")
	PSMDBCmd.AddCommand(bcpCmd)
}

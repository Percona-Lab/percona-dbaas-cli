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

// restoreCmd represents the list command
var restoreCmd = &cobra.Command{
	Use:   "restore-backup <psmdb-cluster-name> <psmdb-backup-name>",
	Short: "Restore MongoDB backup",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name and psmdb-backup-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		args = parseArgs(args)
		name := args[0]
		bcp, err := psmdb.NewRestore(name, *envBckpCrt)
		if err != nil {
			psmdb.PrintError(*backupRestoreAnswerOutput, "new PSMDB restore", err)
			return
		}

		if len(args) < 2 || args[1] == "" {
			psmdb.PrintError(*backupRestoreAnswerOutput, "you have to specify psmdb-cluster-name and psmdb-backup-name", nil)
			return
		}
		bcpName := args[1]

		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		sp.Prefix = "Looking for the cluster..."
		sp.FinalMSG = ""
		sp.Start()
		defer sp.Stop()

		ext, err := bcp.Cmd.IsObjExists("psmdb", name)
		if err != nil {
			psmdb.PrintError(*backupRestoreAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			psmdb.PrintError(*backupRestoreAnswerOutput, "Unable to find cluster psmdb/"+name, nil)
			list, err := bcp.Cmd.List("psmdb")
			if err != nil {
				psmdb.PrintError(*backupRestoreAnswerOutput, "psmdb cluster list", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		sp.Prefix = "Looking for the backup..."
		ext, err = bcp.Cmd.IsObjExists("psmdb-backup", bcpName)
		if err != nil {
			psmdb.PrintError(*backupRestoreAnswerOutput, "check if backup exists", err)
			return
		}

		if !ext {
			sp.Stop()
			fmt.Fprintf(os.Stderr, "Unable to find backup \"%s/%s\"\n", "psmdb-backup", bcpName)
			list, err := bcp.Cmd.List("psmdb-backup")
			if err != nil {
				psmdb.PrintError(*backupRestoreAnswerOutput, "psmdb cluster list", err)
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

		ok := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		go bcp.Create(ok, msg, cerr)
		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for {
			select {
			case okmsg := <-ok:
				sp.FinalMSG = fmt.Sprintf("Restoring backup...[done]\n%s\n", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					psmdb.PrintError(*backupRestoreAnswerOutput, "operator log error", fmt.Errorf(omsg.String()))
					sp.Start()
				}
			case err := <-cerr:
				psmdb.PrintError(*backupRestoreAnswerOutput, "psmdb cluster list", err)
				return
			}
		}
	},
}
var envBckpRstr *string
var backupRestoreAnswerOutput *string

func init() {
	envBckpRstr = restoreCmd.Flags().String("environment", "", "Target kubernetes cluster")
	backupRestoreAnswerOutput = restoreCmd.Flags().String("output", "", "Output format")

	PSMDBCmd.AddCommand(restoreCmd)
}

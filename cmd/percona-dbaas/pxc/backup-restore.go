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
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
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
			log.Error("you have to specify pxc-cluster-name and pxc-backup-name")
			return
		}
		bcpName := args[1]

		dbservice, err := dbaas.New(*envBckpRstr)
		if err != nil {
			log.Error("new dbservice:", err)
			return
		}
		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		sp.Prefix = "Looking for the cluster..."
		sp.FinalMSG = ""
		sp.Start()
		defer sp.Stop()

		ext, err := dbservice.IsObjExists("pxc", name)

		if err != nil {
			log.Error("check if cluster exists:", err)
			return
		}

		if !ext {
			sp.Stop()
			log.Error("unable to find cluster pxc/" + bcpName)
			list, err := dbservice.List("pxc")
			if err != nil {
				log.Error("check if clusters list:", err)
				return
			}
			log.Error("avaliable clusters:", list)
			return
		}

		sp.Prefix = "Looking for the backup..."
		ext, err = dbservice.IsObjExists("pxc-backup", bcpName)
		if err != nil {
			log.Error("check if backup exists:", err)
			return
		}

		if !ext {
			sp.Stop()
			log.Error("unable to find backup pxc-backup/" + bcpName)
			list, err := dbservice.List("pxc-backup")
			if err != nil {
				log.Error("new dbservices", err)
				return
			}
			log.Println("avaliable backups", list)
			return
		}
		sp.Lock()
		sp.Prefix = "Restoring backup..."
		sp.Unlock()
		bcp := pxc.NewRestore(name)

		bcp.Setup(bcpName)

		ok := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		go dbservice.ApplyCheck("pxc-restore", bcp, ok, msg, cerr)
		tckr := time.NewTicker(1 * time.Second)
		defer tckr.Stop()
		for {
			select {
			case okmsg := <-ok:
				sp.FinalMSG = ""
				sp.Stop()
				log.Println("Restoring backup done.", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					log.Error("operator log error", err)
					sp.Start()
				}
			case err := <-cerr:
				log.Error("restore backup:", err)
				return
			}
		}
	},
}

var envBckpRstr *string
var backupRestoreAnswerFormat *string

func init() {
	envBckpRstr = restoreCmd.Flags().String("environment", "", "Target kubernetes cluster")
	backupRestoreAnswerFormat = restoreCmd.Flags().String("output", "", "Answers format")

	PXCCmd.AddCommand(restoreCmd)
}

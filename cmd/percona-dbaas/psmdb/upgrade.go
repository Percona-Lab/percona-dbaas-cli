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
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
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
		dbservice, err := dbaas.New(*envUpgrd)
		if err != nil {
			log.Error("new dbservice:", err)
			return
		}

		app := psmdb.New(name, "doesnotMatter", defaultVersion, "")

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

		ext, err := dbservice.IsObjExists("psmdb", name)
		if err != nil {
			log.Error("check if cluster exists:", err)
			return
		}

		if !ext {
			sp.Stop()
			log.Error("unable to find cluster psmdb/" + name)
			list, err := dbservice.List("psmdb")
			if err != nil {
				log.Error("list psmdb clusters:", err)
				return
			}
			log.Println("avaliable clusters:", list)
			return
		}

		created := make(chan dbaas.Msg)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		oparg := ""
		if len(args) > 1 {
			oparg = args[1]
		}
		appsImg, err := app.Images(oparg, cmd.Flags())
		if err != nil {
			log.Error("setup images for upgrade:", err)
			return
		}

		go dbservice.Upgrade("psmdb", app, appsImg, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Upgrading cluster..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := dbservice.ListName("psmdb", name)
				sp.FinalMSG = ""
				sp.Stop()
				log.Println("upgrade cluster done.", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					log.Error("perator log error:", omsg.String())
					sp.Start()
				}
			case err := <-cerr:
				log.Error("upgrade psmdb:", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envUpgrd *string
var upgradeAnswerFormat *string

func init() {
	upgradeCmd.Flags().String("database-image", "", "Custom image to upgrade psmdb to")
	upgradeCmd.Flags().String("backup-image", "", "Custom image to upgrade backup to")
	envUpgrd = upgradeCmd.Flags().String("environment", "", "Target kubernetes cluster")

	upgradeAnswerFormat = upgradeCmd.Flags().String("output", "", "Answers format")

	PSMDBCmd.AddCommand(upgradeCmd)
}

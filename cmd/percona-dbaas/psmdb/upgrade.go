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
			if *upgradeAnswerInJSON {
				fmt.Fprint(os.Stderr, psmdb.JSONErrorMsg("new dbservice", err))
				return
			}
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			return
		}

		app := psmdb.New(name, "doesnotMatter", defaultVersion, *upgradeAnswerInJSON)

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
			if *upgradeAnswerInJSON {
				fmt.Fprint(os.Stderr, psmdb.JSONErrorMsg("check if cluster exists", err))
				return
			}
			fmt.Fprintf(os.Stderr, "[ERROR] check if cluster exists: %v\n", err)
			return
		}

		if !ext {
			sp.Stop()
			fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "psmdb", name)
			list, err := dbservice.List("psmdb")
			if err != nil {
				if *upgradeAnswerInJSON {
					fmt.Fprint(os.Stderr, psmdb.JSONErrorMsg("psmdb cluster list", err))
					return
				}
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
			if *upgradeAnswerInJSON {
				fmt.Fprint(os.Stderr, psmdb.JSONErrorMsg("setup images for upgrade", err))
				return
			}
			fmt.Fprintf(os.Stderr, "[ERROR] setup images for upgrade: %v\n", err)
			return
		}

		go dbservice.Upgrade("psmdb", app, appsImg, created, msg, cerr)

		sp.Prefix = "Upgrading cluster..."

		for {
			select {
			case <-created:
				okmsg, _ := dbservice.ListName("psmdb", name)
				sp.FinalMSG = fmt.Sprintf("Upgrading cluster...[done]\n\n%s", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					if *upgradeAnswerInJSON {
						fmt.Fprint(os.Stderr, psmdb.JSONErrorMsg("operator log error", fmt.Errorf(omsg.String())))
					} else {
						fmt.Printf("[operator log error] %s\n", omsg)
					}
					sp.Start()
				}
			case err := <-cerr:
				if *upgradeAnswerInJSON {
					fmt.Fprint(os.Stderr, psmdb.JSONErrorMsg("upgrade psmdb", err))
					return
				}
				fmt.Fprintf(os.Stderr, "\n[ERROR] upgrade psmdb: %v\n", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envUpgrd *string
var upgradeAnswerInJSON *bool

func init() {
	upgradeCmd.Flags().String("database-image", "", "Custom image to upgrade psmdb to")
	upgradeCmd.Flags().String("backup-image", "", "Custom image to upgrade backup to")
	envUpgrd = upgradeCmd.Flags().String("environment", "", "Target kubernetes cluster")

	upgradeAnswerInJSON = upgradeCmd.Flags().Bool("json", false, "Answers in JSON format")

	PSMDBCmd.AddCommand(upgradeCmd)
}

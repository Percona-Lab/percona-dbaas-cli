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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
)

// upgradeOperatorCmd represents the edit command
var upgradeOperatorCmd = &cobra.Command{
	Use:   "upgrade-operator <psmdb-cluster-name> <to-version>",
	Short: "Upgrade PSMDB operator",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		switch *upgradeOperatorAnswerFormat {
		case "json":
			log.Formatter = new(logrus.JSONFormatter)
		}
		dbservice, err := dbaas.New(*envUpgrdOprtr)
		if err != nil {
			log.Errorln("new dbservice:", err.Error())
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
			log.Errorln("check if cluster exists:", err.Error())
			return

		}

		if !ext {
			sp.Stop()
			log.Println("unable to find cluster psmdb/", name)
			list, err := dbservice.List("psmdb")
			if err != nil {
				log.Errorln("db service list:", err.Error())
				return
			}
			log.Println("avaliable clusters:", list)
			return
		}

		created := make(chan string)
		cerr := make(chan error)

		if *oprtrImage != "" {
			num, err := dbservice.Instances("psmdb")
			if err != nil {
				log.Errorln("unable to get psmdb instances:", err.Error())
				return
			}
			if len(num) > 1 {
				sp.Stop()
				var yn string
				fmt.Printf("\nFound more than one psmdb cluster: %v.\nOperator upgrade may affect other clusters.\nContinue? [y/N] ", num)
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

		go dbservice.UpgradeOperator(app, *oprtrImage, created, cerr)
		sp.Lock()
		sp.Prefix = "Upgrading cluster operator..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := dbservice.ListName("psmdb", name)
				sp.FinalMSG = ""
				sp.Stop()
				log.Println("upgrading cluster operator done.", okmsg)
				return
			case err := <-cerr:
				log.Errorln("upgrade psmdb operator:", err.Error())
				sp.HideCursor = true
				return
			}
		}
	},
}

var envUpgrdOprtr *string
var oprtrImage *string
var upgradeOperatorAnswerFormat *string

func init() {
	oprtrImage = upgradeOperatorCmd.Flags().String("operator-image", "", "Custom image to upgrade operator to")
	envUpgrdOprtr = upgradeOperatorCmd.Flags().String("environment", "", "Target kubernetes cluster")
	upgradeOperatorAnswerFormat = upgradeOperatorCmd.Flags().String("output", "", "Answers format")

	PSMDBCmd.AddCommand(upgradeOperatorCmd)
}

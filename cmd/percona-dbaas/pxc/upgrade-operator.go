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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

// upgradeOperatorCmd represents the edit command
var upgradeOperatorCmd = &cobra.Command{
	Use:   "upgrade-operator <pxc-cluster-name> <to-version>",
	Short: "Upgrade PXC operator",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("you have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		dbservice, err := dbaas.New(*envUpgrdOprtr)
		if err != nil {
			log.Error("new dbservice: ", err)
			return
		}
		app := pxc.New(name, defaultVersion, "")

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

		ext, err := dbservice.IsObjExists("pxc", name)

		if err != nil {
			log.Error("check if cluster exists: ", err)
			return
		}

		if !ext {
			sp.Stop()
			log.Errorf("unable to find cluster \"%s/%s\"\n", "pxc", name)
			list, err := dbservice.List("pxc")
			if err != nil {
				log.Error("cluster list: ", err)
				return
			}

			log.Println("Avaliable clusters\n", list)
			return
		}

		created := make(chan dbaas.Msg)
		cerr := make(chan error)

		if *oprtrImage != "" {
			num, err := dbservice.Instances("pxc")
			if err != nil {
				log.Error("unable to get pxc instances: ", err)
			}
			if len(num) > 1 {
				sp.Stop()
				var yn string
				fmt.Printf("\nFound more than one pxc cluster: %s.\nOperator upgrade may affect other clusters.\nContinue? [y/N] ", num)
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
				okmsg, _ := dbservice.ListName("pxc", name)
				sp.FinalMSG = ""
				sp.Stop()
				log.WithField("data", okmsg).Info("upgrading cluster operator done.")
				return
			case err := <-cerr:
				log.Error("upgrade pxc operator: ", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envUpgrdOprtr *string
var oprtrImage *string

func init() {
	oprtrImage = upgradeOperatorCmd.Flags().String("operator-image", "", "Custom image to upgrade operator to")
	envUpgrdOprtr = upgradeOperatorCmd.Flags().String("environment", "", "Target kubernetes cluster")

	PXCCmd.AddCommand(upgradeOperatorCmd)
}

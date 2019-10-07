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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "modify-db <psmdb-cluster-name>",
	Short: "Modify MongoDB cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		args = parseArgs(args)
		switch *editAnswerFormat {
		case "json":
			log.Formatter = new(logrus.JSONFormatter)
		}
		clusterName := args[0]

		dbservice, err := dbaas.New(*envEdt)
		if err != nil {
			log.Errorln("new dbservice:", err.Error())
			return
		}
		rsName := ""
		if len(args) >= 2 {
			rsName = args[1]
		}

		app := psmdb.New(clusterName, rsName, defaultVersion, "")

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

		ext, err := dbservice.IsObjExists("psmdb", clusterName)
		if err != nil {
			log.Errorln("check if cluster exists:", err.Error())
			return
		}

		if !ext {
			sp.Stop()
			log.Errorln("unable to find cluster psmdb/" + clusterName)
			list, err := dbservice.List("psmdb")
			if err != nil {
				log.Errorln("psmdb cluster list:", err.Error())
				return
			}
			log.Println("avaliable clusters:", list)
			return
		}

		created := make(chan dbaas.Msg)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)
		config, err := psmdb.ParseEditFlagsToConfig(cmd.Flags())
		if err != nil {
			log.Errorln("parsing flags:", err.Error())
			return
		}
		app.ClusterConfig = config
		go dbservice.Edit("psmdb", app, nil, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Applying changes..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := dbservice.ListName("psmdb", clusterName)
				sp.FinalMSG = ""
				sp.Stop()
				log.Println("aApplying changes done", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					log.Errorln("operator log error:", omsg.String())
					sp.Start()
				}
			case err := <-cerr:
				log.Errorln("edit psmdb:", err.Error())
				sp.HideCursor = true
				return
			}
		}
	},
}

var envEdt *string
var editAnswerFormat *string

func init() {
	editCmd.Flags().Int32("replset-size", 3, "Number of nodes in replset")
	envEdt = editCmd.Flags().String("environment", "", "Target kubernetes cluster")
	editAnswerFormat = editCmd.Flags().String("output", "", "Answers format")

	PSMDBCmd.AddCommand(editCmd)
}

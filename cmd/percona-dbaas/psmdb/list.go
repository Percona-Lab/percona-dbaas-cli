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
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "describe-db <db-cluster-name>",
	Short: "Show either specific MongoDB cluster or all clusters on current Kubernetes environment",
	Run: func(cmd *cobra.Command, args []string) {
		dbservice, err := dbaas.New(*envLst)
		if err != nil {
			log.Error("new dbservice: ", err)
			return
		}
		if len(args) > 0 {
			app := psmdb.New(args[0], "", defaultVersion, "")
			info, err := dbservice.Describe(app)
			if err != nil {
				log.Error("describe: ", err)
				return
			}
			log.WithField("data", info).Info("Cluster data")
			return
		}
		list, err := dbservice.List("psmdb")
		if err != nil {
			log.Error("list psmdb clusters: ", err)
			return
		}

		log.WithField("data", list).Info("List")
	},
}

var envLst *string

func init() {
	envLst = listCmd.Flags().String("environment", "", "Target kubernetes cluster")

	PSMDBCmd.AddCommand(listCmd)
}

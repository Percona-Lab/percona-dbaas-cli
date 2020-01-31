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

package mysql

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	dbaas "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe-db <mysql-cluster-name>",
	Short: "Describe MySQL cluster or list clusters",
	Long:  "Lists all database instances or clusters currently present or provides details about the database instance or cluster with the given name.",
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		instance := dbaas.Instance{
			Name:          name,
			EngineOptions: *options,
			Engine:        *descrEngine,
			Provider:      *descrProvider,
		}

		if len(name) > 0 {
			db, err := dbaas.DescribeDB(instance)
			if err != nil {
				log.Error("describe db: ", err)
				return
			}
			db.Pass = ""
			log.Print(db)
			return
		}

		listDB, err := dbaas.ListDB(instance)
		if err != nil {
			log.Error("list db: ", err)
			return
		}
		if len(listDB) == 0 {
			log.Println("Nothing to show")
			return
		}
		log.Println("NAME                STATUS")
		for _, db := range listDB {
			space := "          "
			if len(db.ResourceName) <= 12 {
				for i := 0; i < 12-len(db.ResourceName); i++ {
					space = space + " "
				}
			} else {
				db.ResourceName = db.ResourceName[:12]
			}
			log.Println(db.ResourceName + space + db.Status)
		}
	},
}

var descrProvider *string
var descrEngine *string

func init() {
	descrProvider = describeCmd.Flags().String("provider", "k8s", "Provider")
	descrEngine = describeCmd.Flags().String("engine", "pxc", "Engine")

	PXCCmd.AddCommand(describeCmd)
}

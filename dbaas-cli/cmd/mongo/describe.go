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

package mongo

import (
	"fmt"
	"os"
	"text/tabwriter"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/cmd/tools"
	dbaas "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"
)

// describeCmd represents the describe command
var describeCmd = &cobra.Command{
	Use:   "describe-db <mongo-cluster-name>",
	Short: "Describe MongoDB cluster or list clusters",
	Long:  "Lists all database instances or clusters currently present or provides details about the database instance or cluster with the given name.",
	Run: func(cmd *cobra.Command, args []string) {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		instance := tools.GetInstance(name, "", *descrEngine, *descrProvider, "")

		if len(name) > 0 {
			db, err := dbaas.DescribeDB(instance)
			if err != nil {
				log.Error("describe db: ", err)
				return
			}
			db.Pass = ""
			log.WithField("database", db).Info("information")
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

		format, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Error("get output flag: ", err)
			return
		}
		switch format {
		case "json":
			log.WithField("database-list", listDB).Info("information")
		default:
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 8, 0, '\t', 0)
			fmt.Fprintln(w, "NAME\tSTATUS\t")
			for _, db := range listDB {
				fmt.Fprintln(w, fmt.Sprintf("%s\t%s", db.ResourceName, db.Status))

			}
			fmt.Fprintln(w)
			w.Flush()
		}
	},
}

var descrProvider *string
var descrEngine *string

func init() {
	descrProvider = describeCmd.Flags().String("provider", "k8s", "Provider")
	descrEngine = describeCmd.Flags().String("engine", "psmdb", "Engine")

	MongoCmd.AddCommand(describeCmd)
}

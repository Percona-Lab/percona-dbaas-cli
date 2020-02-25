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
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/dp"
	dbaas "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"
	_ "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb"
)

const (
	defaultVersion = "default"
	maxTries       = 1200
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create-db <mongo-cluster-name>",
	Short: "Create MongoDB cluster",
	Long:  "Creates a new databases instance or cluster with the given name.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify resource name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		noWait, err := cmd.Flags().GetBool("no-wait")
		if err != nil {
			log.Error("get no-wait flag: ", err)
		}
		output, err := cmd.Flags().GetString("output")
		if err != nil {
			log.Error("get output flag: ", err)
		}
		dotPrinterInBackground := false
		if output == "json" {
			dotPrinterInBackground = true
		}
		dotPrinter := dp.New(dotPrinterInBackground)
		if len(*options) > 0 {
			*options = addSpec(*options)
		}
		instance := dbaas.Instance{
			Name:          args[0],
			EngineOptions: *options,
			Engine:        *engine,
			Provider:      *provider,
			RootPass:      *rootPass,
		}
		dotPrinter.Start("Starting")
		err = dbaas.CreateDB(instance)
		if err != nil {
			dotPrinter.Stop("error")
			log.Error("create db: ", err)
			return
		}
		tries := 0
		tckr := time.NewTicker(500 * time.Millisecond)
		defer tckr.Stop()
		for range tckr.C {
			cluster, err := dbaas.DescribeDB(instance)
			if err != nil {
				//dotPrinter.StopPrintDot("error")
				//log.Error("check db: ", err)
				continue
			}
			switch cluster.Status {
			case "ready":
				dotPrinter.Stop("done")
				cluster.Message = strings.Replace(cluster.Message, "PASSWORD", cluster.Pass, 1)
				log.WithField("database", cluster).Info("Database started successfully, connection details are below:")
				return
			case "initializing":
				if noWait {
					dotPrinter.Stop("initializing")
					cluster.Message = strings.Replace(cluster.Message, "PASSWORD", cluster.Pass, 1)
					log.WithField("database", cluster).Info("information")
					return
				}
			}
			if tries >= maxTries {
				dotPrinter.Stop("error")
				log.Error("unable to start cluster. cluster status: ", cluster.Status)
				return
			}
			tries++
		}
	},
}

var options *string
var provider *string
var engine *string
var rootPass *string

func init() {
	options = createCmd.Flags().String("options", "", "Engine options in 'p1.p2=text' format. For k8s/psmdb use params from https://www.percona.com/doc/kubernetes-operator-for-psmongodb/operator.html")
	provider = createCmd.Flags().String("provider", "k8s", "Provider")
	engine = createCmd.Flags().String("engine", "psmdb", "Engine")
	rootPass = createCmd.Flags().String("password", "", "Password for superuser")

	MongoCmd.AddCommand(createCmd)
}

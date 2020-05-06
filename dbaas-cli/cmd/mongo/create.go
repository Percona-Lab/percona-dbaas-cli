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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-cli/cmd/tools"
	dbaas "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"
	_ "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-psmdb"
)

const (
	defaultVersion = "default"
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
		err := setupOutput(cmd)
		if err != nil {
			log.Error(err)
			return
		}
		instance := tools.GetInstance(args[0], addSpec(*options), *engine, *provider, *rootPass)

		warns, err := dbaas.PreCheck(instance)
		for _, w := range warns {
			log.Println("Warning:", w)
		}
		if err != nil {
			log.Error(err)
			return
		}

		dotPrinter.Start("Starting")
		err = dbaas.CreateDB(instance)
		if err != nil {
			dotPrinter.Stop("error")
			log.Error("create db: ", err)
			return
		}
		cluster, err := tools.GetDB(instance, false, noWait, maxTries)
		if err != nil {
			dotPrinter.Stop("error")
			log.Errorf("unable to start cluster:", err)
			return
		}

		if cluster.Status == dbaas.StateInit {
			dotPrinter.Stop("initializing")
			log.WithField("database", cluster).Info("information")
			return
		}

		dotPrinter.Stop("done")
		log.WithField("database", cluster).Info("Database started successfully, connection details are below:")
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

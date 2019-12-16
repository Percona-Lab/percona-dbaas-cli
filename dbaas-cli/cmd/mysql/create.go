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
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	dbaas "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib"
	_ "github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc"
)

const (
	defaultVersion = "default"
	maxTries       = 1200
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create-db <mysql-cluster-name>",
	Short: "Create MySQL cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify resource name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		instance := dbaas.Instance{
			Name:          args[0],
			EngineOptions: *options,
			Engine:        *engine,
			Provider:      *provider,
		}

		dotPrinter.StartPrintDot("Starting")
		err := dbaas.CreateDB(instance)
		if err != nil {
			//sp.Stop()
			dotPrinter.StopPrintDot("error")
			log.Error("create db: ", err)
			return
		}
		tries := 0
		tckr := time.NewTicker(500 * time.Millisecond)
		defer tckr.Stop()
		for range tckr.C {
			cluster, err := dbaas.DescribeDB(instance)
			if err != nil {
				//sp.Stop()
				dotPrinter.StopPrintDot("error")
				log.Error("check db: ", err)
				return
			}
			if cluster.Status == "ready" {
				//sp.Stop()
				dotPrinter.StopPrintDot("done")
				log.Println("Database started successfully, connection details are below:")
				log.Println(cluster)
				return
			}
			if tries >= maxTries {
				//sp.Stop()
				dotPrinter.StopPrintDot("error")
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

func init() {
	options = createCmd.Flags().String("options", "", "Engine options")
	provider = createCmd.Flags().String("provider", "", "Provider")
	engine = createCmd.Flags().String("engine", "", "Engine")

	PXCCmd.AddCommand(createCmd)
}

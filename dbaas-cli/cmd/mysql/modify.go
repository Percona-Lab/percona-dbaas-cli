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

// modifyCmd represents the create command
var modifyCmd = &cobra.Command{
	Use:   "modify-db <mysql-cluster-name>",
	Short: "Modify MySQL cluster ",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify resource name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		instance := dbaas.Instance{
			Name:          args[0],
			EngineOptions: *modifyOptions,
			Engine:        *modifyEngine,
			Provider:      *modifyProvider,
		}

		dotPrinter.StartPrintDot("Modifying")
		err := dbaas.ModifyDB(instance)
		if err != nil {
			dotPrinter.StopPrintDot("error")
			log.Error("modify db: ", err)
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
			if cluster.Status == "ready" {
				dotPrinter.StopPrintDot("done")
				log.Println("Database modifyed successfully, connection details are below:")
				log.Println(cluster)
				return
			}
			if tries >= maxTries {
				dotPrinter.StopPrintDot("error")
				log.Error("unable to modify cluster. cluster status: ", cluster.Status)
				return
			}
			tries++
		}
	},
}

var modifyOptions *string
var modifyProvider *string
var modifyEngine *string

func init() {
	modifyOptions = modifyCmd.Flags().String("options", "", "Engine options")
	modifyProvider = modifyCmd.Flags().String("provider", "", "Provider")
	modifyEngine = modifyCmd.Flags().String("engine", "", "Engine")

	PXCCmd.AddCommand(modifyCmd)
}

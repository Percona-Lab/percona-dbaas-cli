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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var delePVC *bool

// delCmd represents the list command
var delCmd = &cobra.Command{
	Use:   "delete-db <pxc-cluster-name>",
	Short: "Delete MySQL cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("you have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		/*name := args[0]
		labelsMap := make(map[string]string)
		if len(*labels) > 0 {
			keyValues := strings.Split(*labels, ",")
			for index := range keyValues {
				itemSlice := strings.Split(keyValues[index], "=")
				labelsMap[itemSlice[0]] = itemSlice[1]
			}
		}
		pxcOperator, err := pxc.NewPXCController(labelsMap, *envDlt)
		if err != nil {
			log.Error("new pxc operator: ", err)
			return
		}
		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		sp.Prefix = "Looking for the cluster..."
		sp.FinalMSG = ""
		sp.Start()
		defer sp.Stop()

		if *delePVC {
			sp.Stop()
			var yn string
			fmt.Printf("\nAll current data on \"%s\" cluster will be destroyed.\nAre you sure? [y/N] ", name)
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
		sp.Lock()
		sp.Prefix = "Deleting..."
		sp.Unlock()

		err = pxcOperator.DeleteDBCluster(name, *delePVC)
		if err != nil {
			log.Error("delete cluster: ", err)
			return
		}
		sp.Stop()*/
		log.Println("Deleting done")
	},
}

var envDlt *string

func init() {
	delePVC = delCmd.Flags().Bool("clear-data", false, "Remove cluster volumes")
	envDlt = delCmd.Flags().String("environment", "", "Target kubernetes cluster")

	PXCCmd.AddCommand(delCmd)
}

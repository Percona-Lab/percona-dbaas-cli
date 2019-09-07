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
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const noS3backupOpts = `[Error] S3 backup storage options doesn't set properly: %v.`

// storageCmd represents the edit command
var storageCmd = &cobra.Command{
	Use:   "create-backup-storage <psmdb-cluster-name>",
	Short: "Add storage for MongoDB backups",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		args = parseArgs(args)

		clusterName := args[0]

		rsName := ""
		if len(args) >= 2 {
			rsName = args[1]
		}

		app, err := psmdb.New(clusterName, rsName, defaultVersion, "", *envStor)
		if err != nil {
			psmdb.PrintError(*addStorageAnswerOutput, "create psmdb", err)
			return
		}
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

		ext, err := app.Cmd.IsObjExists("psmdb", clusterName)
		if err != nil {
			psmdb.PrintError(*addStorageAnswerOutput, "check if cluster exists", err)
			return
		}

		if !ext {
			sp.Stop()
			fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "psmdb", clusterName)
			list, err := app.Cmd.List("psmdb")
			if err != nil {
				psmdb.PrintError(*addStorageAnswerOutput, "psmdb list", err)
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		config, err := psmdb.ParseAddStorageFlagsToConfig(cmd.Flags())
		if err != nil {
			psmdb.PrintError(*addStorageAnswerOutput, "parsing flags", err)
			return
		}

		s3stor, err := app.Cmd.S3Storage(app.Name(), config.S3)
		if err != nil {
			switch err.(type) {
			case dbaas.ErrNoS3Options:
				psmdb.PrintError(*addStorageAnswerOutput, "S3 backup storage options doesn't set properly", err)
			default:
				psmdb.PrintError(*addStorageAnswerOutput, "create S3 backup storage", err)
			}
			return
		}

		created := make(chan psmdb.ClusterData)
		msg := make(chan psmdb.ClusterData)
		cerr := make(chan error)
		app.ClusterConfig = config
		go app.Edit(s3stor, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Adding the storage..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := app.Cmd.ListName("psmdb", clusterName)
				finalMsg, err := SprintResponse(*addStorageAnswerOutput, okmsg)
				if err != nil {
					pxc.PrintError(*addStorageAnswerOutput, "sprint response", err)
				}
				sp.FinalMSG = fmt.Sprintln("Adding the storage...[done]\n\n", finalMsg)
				return
			case omsg := <-msg:
				sp.Stop()
				psmdb.PrintError(*addStorageAnswerOutput, "operator log error", fmt.Errorf(omsg.Message))
				sp.Start()
			case err := <-cerr:
				psmdb.PrintError(*addStorageAnswerOutput, "add storage to psmdb", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envStor *string
var addStorageAnswerOutput *string

func init() {
	storageCmd.Flags().String("s3-endpoint-url", "", "Endpoing URL of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-bucket", "", "Bucket of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-region", "", "Region of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-credentials-secret", "", "Secrets with credentials for S3 compatible storage to store backup at. Alternatevily you can set --s3-access-key-id and --s3-secret-access-key instead.")
	storageCmd.Flags().String("s3-access-key-id", "", "Access Key ID for S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-secret-access-key", "", "Access Key for S3 compatible storage to store backup at")
	envStor = storageCmd.Flags().String("environment", "", "Target kubernetes cluster")

	storageCmd.Flags().Int32("replset-size", 0, "Number of nodes in replset")

	addStorageAnswerOutput = storageCmd.Flags().String("output", "", "Output format")

	PSMDBCmd.AddCommand(storageCmd)
}

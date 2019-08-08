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
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const noS3backupOpts = `[Error] S3 backup storage options doesn't set properly: %v.`

// storageCmd represents the edit command
var storageCmd = &cobra.Command{
	Use:   "create-backup-storage <pxc-cluster-name>",
	Short: "Add storage for MySQL backups",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		clusterName := args[0]

		dbgeneric, err := dbaas.New(*envStor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)
			return
		}
		app := pxc.New(clusterName, defaultVersion, *dbgeneric)

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

		ext, err := dbgeneric.IsObjExists("pxc", clusterName)

		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] check if cluster exists: %v\n", err)
			return
		}

		if !ext {
			sp.Stop()
			fmt.Fprintf(os.Stderr, "Unable to find cluster \"%s/%s\"\n", "pxc", clusterName)
			list, err := dbgeneric.List("pxc")
			if err != nil {
				return
			}
			fmt.Println("Avaliable clusters:")
			fmt.Print(list)
			return
		}

		config, err := pxc.ParseAddStorageFlagsToConfig(cmd.Flags())
		if err != nil {
			fmt.Println("[Error] parse flags to config:", err)
			return
		}

		s3stor, err := dbgeneric.S3Storage(app, config.S3)
		if err != nil {
			switch err.(type) {
			case dbaas.ErrNoS3Options:
				fmt.Printf(noS3backupOpts, err)
			default:
				fmt.Println("[Error] create S3 backup storage:", err)
			}
			return
		}

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		go dbgeneric.Edit("pxc", app, config, s3stor, created, msg, cerr)
		sp.Prefix = "Adding the storage..."

		for {
			select {
			case <-created:
				okmsg, _ := dbgeneric.ListName("pxc", clusterName)
				sp.FinalMSG = fmt.Sprintf("Adding the storage...[done]\n\n%s", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					fmt.Printf("[operator log error] %s\n", omsg)
					sp.Start()
				}
			case err := <-cerr:
				fmt.Fprintf(os.Stderr, "\n[ERROR] add storage to pxc: %v\n", err)
				sp.HideCursor = true
				return
			}
		}
	},
}

var envStor *string

func init() {
	storageCmd.Flags().String("s3-endpoint-url", "", "Endpoing URL of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-bucket", "", "Bucket of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-region", "", "Region of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-credentials-secret", "", "Secrets with credentials for S3 compatible storage to store backup at. Alternatevily you can set --s3-access-key-id and --s3-secret-access-key instead.")
	storageCmd.Flags().String("s3-access-key-id", "", "Access Key ID for S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-secret-access-key", "", "Access Key for S3 compatible storage to store backup at")

	storageCmd.Flags().Int32("pxc-instances", 0, "Number of PXC nodes in cluster")
	storageCmd.Flags().Int32("proxy-instances", 0, "Number of ProxySQL nodes in cluster")
	envStor = storageCmd.Flags().String("environment", "", "Target kubernetes cluster")

	PXCCmd.AddCommand(storageCmd)
}

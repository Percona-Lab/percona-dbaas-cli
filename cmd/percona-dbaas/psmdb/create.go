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
)

const (
	defaultVersion = "1.0.0"

	noS3backupWarn = `[Error] S3 backup storage options doesn't set: %v. You have specify S3 storage in order to make backups.
You can skip this step by using --s3-skip-storage flag add the storage later with the "add-storage" command.
`
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create <psmdb-cluster-name> <replica-set-name>",
	Short: "Create MongoDB cluster on current Kubernetes cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify psmdb-cluster-name")
		}

		if len(args) > 2 {
			return errors.Errorf("Unknow arguments %v", args[2:])
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

		app := psmdb.New(clusterName, rsName, defaultVersion)

		var s3stor *dbaas.BackupStorageSpec
		if !*skipS3Storage {
			var err error
			s3stor, err = dbaas.S3Storage(app, cmd.Flags())
			if err != nil {
				switch err.(type) {
				case dbaas.ErrNoS3Options:
					fmt.Printf(noS3backupWarn, err)
				default:
					fmt.Println("[Error] create S3 backup storage:", err)
				}
				return
			}
		}

		setupmsg, err := app.Setup(cmd.Flags(), s3stor)
		if err != nil {
			fmt.Println("[Error] set configuration:", err)
			return
		}

		fmt.Println(setupmsg)

		created := make(chan string)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		go dbaas.Create("psmdb", app, created, msg, cerr)
		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		demo, err := cmd.Flags().GetBool("demo")
		if demo && err == nil {
			sp.UpdateCharSet([]string{""})
		}
		sp.Prefix = "Starting..."
		sp.Start()
		defer sp.Stop()
		for {
			select {
			case okmsg := <-created:
				sp.FinalMSG = fmt.Sprintf("Starting...[done]\n%s\n", okmsg)
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
				sp.Stop()
				switch err.(type) {
				case dbaas.ErrAlreadyExists:
					fmt.Fprintf(os.Stderr, "\n[ERROR] %v\n", err)
					list, err := dbaas.List("psmdb")
					if err != nil {
						return
					}
					fmt.Println("Avaliable clusters:")
					fmt.Print(list)
				default:
					fmt.Fprintf(os.Stderr, "\n[ERROR] create psmdb: %v\n", err)
				}

				return
			}
		}
	},
}

var skipS3Storage *bool

func init() {
	createCmd.Flags().String("storage-size", "6G", "Node volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)")
	createCmd.Flags().String("storage-class", "", "Name of the StorageClass required by the volume claim")
	createCmd.Flags().Int32("replset-size", 3, "Number of nodes in replset")
	createCmd.Flags().String("request-cpu", "600m", "Node requests for CPU, in cores. (500m = .5 cores)")
	createCmd.Flags().String("request-mem", "1G", "Node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	createCmd.Flags().String("anti-affinity-key", "kubernetes.io/hostname", "Pod anti-affinity rules. Allowed values: none, kubernetes.io/hostname, failure-domain.beta.kubernetes.io/zone, failure-domain.beta.kubernetes.io/region")

	createCmd.Flags().String("s3-endpoint-url", "", "Endpoing URL of S3 compatible storage to store backup at")
	createCmd.Flags().String("s3-bucket", "", "Bucket of S3 compatible storage to store backup at")
	createCmd.Flags().String("s3-region", "", "Region of S3 compatible storage to store backup at")
	createCmd.Flags().String("s3-credentials-secret", "", "Secrets with credentials for S3 compatible storage to store backup at. Alternatevily you can set --s3-key-id and --s3-key instead.")
	createCmd.Flags().String("s3-key-id", "", "Access Key ID for S3 compatible storage to store backup at")
	createCmd.Flags().String("s3-key", "", "Access Key for S3 compatible storage to store backup at")
	skipS3Storage = createCmd.Flags().Bool("s3-skip-storage", false, "Don't create S3 compatible backup storage. Has to be set manually later on.")

	PSMDBCmd.AddCommand(createCmd)
}

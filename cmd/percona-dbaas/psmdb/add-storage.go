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
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
)

const noS3backupOpts = `S3 backup storage options doesn't set properly: %v.`

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
		switch *addStorageAnswerFormat {
		case "json":
			log.Formatter = new(logrus.JSONFormatter)
		}
		clusterName := args[0]
		dbservice, err := dbaas.New(*envStor)
		if err != nil {
			log.Errorln("new dbservice:", err.Error())
			return
		}
		rsName := ""
		if len(args) >= 2 {
			rsName = args[1]
		}

		app := psmdb.New(clusterName, rsName, defaultVersion, "")

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

		ext, err := dbservice.IsObjExists("psmdb", clusterName)
		if err != nil {
			log.Errorln("check if cluster exists:", err.Error())
			return
		}

		if !ext {
			sp.Stop()
			log.Errorln("unable to find cluster psmdb/" + clusterName)
			list, err := dbservice.List("psmdb")
			if err != nil {
				log.Errorln("psmdb list:", err.Error())
				return
			}

			log.Println("avaliable clusters:", list)
			return
		}

		config, err := psmdb.ParseAddStorageFlagsToConfig(cmd.Flags())
		if err != nil {
			log.Errorln("parsing flags", err.Error())
		}

		s3stor, err := dbservice.S3Storage(app, config.S3)
		if err != nil {
			switch err.(type) {
			case dbaas.ErrNoS3Options:
				log.Printf(noS3backupOpts, err)
			default:
				log.Println("create S3 backup storage:", err.Error)
			}
			return
		}

		created := make(chan dbaas.Msg)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)
		app.ClusterConfig = config
		go dbservice.Edit("psmdb", app, s3stor, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Adding the storage..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := dbservice.ListName("psmdb", clusterName)
				sp.FinalMSG = ""
				sp.Stop()
				log.Println("adding the storage done.", okmsg)
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					log.Errorln("operator log error:", omsg.String())
					sp.Start()
				}
			case err := <-cerr:
				log.Println("add storage to psmdb:", err.Error())
				sp.HideCursor = true
				return
			}
		}
	},
}

var envStor *string
var addStorageAnswerFormat *string

func init() {
	storageCmd.Flags().String("s3-endpoint-url", "", "Endpoing URL of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-bucket", "", "Bucket of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-region", "", "Region of S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-credentials-secret", "", "Secrets with credentials for S3 compatible storage to store backup at. Alternatevily you can set --s3-access-key-id and --s3-secret-access-key instead.")
	storageCmd.Flags().String("s3-access-key-id", "", "Access Key ID for S3 compatible storage to store backup at")
	storageCmd.Flags().String("s3-secret-access-key", "", "Access Key for S3 compatible storage to store backup at")
	envStor = storageCmd.Flags().String("environment", "", "Target kubernetes cluster")

	storageCmd.Flags().Int32("replset-size", 0, "Number of nodes in replset")

	addStorageAnswerFormat = storageCmd.Flags().String("output", "", "Answers format")

	PSMDBCmd.AddCommand(storageCmd)
}

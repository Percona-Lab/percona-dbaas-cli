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
	"time"

	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const noS3backupOpts = `S3 backup storage options doesn't set properly: %v.`

// storageCmd represents the edit command
var storageCmd = &cobra.Command{
	Use:   "create-backup-storage <pxc-cluster-name>",
	Short: "Add storage for MySQL backups",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("you have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		clusterName := args[0]

		dbservice, err := dbaas.New(*envStor)
		if err != nil {
			log.Error("new dbservice: ", err)
			return
		}

		app := pxc.New(clusterName, defaultVersion, "")

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

		ext, err := dbservice.IsObjExists("pxc", clusterName)
		if err != nil {
			log.Error("check if cluster exists: ", err)
			return
		}

		if !ext {
			sp.Stop()
			log.Errorf("unable to find cluster \"%s/%s\"\n", "pxc", clusterName)
			list, err := dbservice.List("pxc")
			if err != nil {
				log.Error("list pxc clusters: ", err)
				return
			}
			log.Println("Avaliable clusters:\n", list)
			return
		}

		config, err := pxc.ParseAddStorageFlagsToConfig(cmd.Flags())
		if err != nil {
			log.Error("parse flags to config: ", err)
			return
		}
		app.ClusterConfig = config

		s3stor, err := dbservice.S3Storage(app, config.S3)
		if err != nil {
			switch err.(type) {
			case dbaas.ErrNoS3Options:
				log.Error(noS3backupOpts, err)
			default:
				log.Error("create S3 backup storage: ", err)
			}
			return
		}

		created := make(chan dbaas.Msg)
		msg := make(chan dbaas.OutuputMsg)
		cerr := make(chan error)

		go dbservice.Edit("pxc", app, s3stor, created, msg, cerr)
		sp.Lock()
		sp.Prefix = "Adding the storage..."
		sp.Unlock()
		for {
			select {
			case <-created:
				okmsg, _ := dbservice.ListName("pxc", clusterName)
				sp.FinalMSG = ""
				sp.Stop()
				log.WithField("data", okmsg).Info("adding the storage...[done]\n\n%s")
				return
			case omsg := <-msg:
				switch omsg.(type) {
				case dbaas.OutuputMsgDebug:
					// fmt.Printf("\n[debug] %s\n", omsg)
				case dbaas.OutuputMsgError:
					sp.Stop()
					log.Error("operator log error: ", err)
					sp.Start()
				}
			case err := <-cerr:
				log.Error("add storage to pxc: ", err)
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

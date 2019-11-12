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

	"github.com/Percona-Lab/percona-dbaas-cli/operator/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/operator/pxc"
	"github.com/briandowns/spinner"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	defaultVersion = "default"

	noS3backupWarn = `S3 backup storage options doesn't set: %v. You have specify S3 storage in order to make backups.
You can skip this step by using --s3-skip-storage flag add the storage later with the "add-storage" command.
`
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create-db <pxc-cluster-name>",
	Short: "Create MySQL cluster on current Kubernetes cluster",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("You have to specify pxc-cluster-name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		dbOperator, err := pxc.NewController(*labels, *envCrt, args[0])
		if err != nil {
			log.Error("new pxc operator: ", err)
			return
		}

		config, err := pxc.ParseCreateFlagsToConfig(cmd.Flags())
		if err != nil {
			log.Error("parse flags to config: ", err)
			return
		}

		sp := spinner.New(spinner.CharSets[14], 250*time.Millisecond)
		sp.Color("green", "bold")
		sp.Lock()
		sp.Prefix = "Starting..."
		sp.Unlock()
		sp.Start()
		defer sp.Stop()
		msg, err := dbOperator.CreatDB(config, *skipS3Storage, *operatorImage)
		if err != nil {
			switch err.(type) {
			case dbaas.ErrAlreadyExists:
				log.Error("create pxc cluster: ", err)
				list, err := dbOperator.Cmd.List("pxc")
				if err != nil {
					log.Error("list pxc clusters: ", err)
					return
				}
				log.Println("Avaliable clusters:\n", list)
				return
			default:
				log.Error(errors.Wrap(err, "create DB"))
				return
			}
		}
		sp.Stop()
		log.WithField("data", msg).Info("Cluster ready")

		return
	},
}

var skipS3Storage *bool
var envCrt *string
var labels *string
var operatorImage *string

func init() {
	createCmd.Flags().String("storage-size", "6G", "PXC node volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)")
	createCmd.Flags().String("storage-class", "", "Name of the StorageClass required by the volume claim")
	createCmd.Flags().Int32("pxc-instances", 3, "Number of PXC nodes in cluster")
	createCmd.Flags().String("pxc-request-cpu", "600m", "PXC node requests for CPU, in cores. (500m = .5 cores)")
	createCmd.Flags().String("pxc-request-mem", "1G", "PXC node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	createCmd.Flags().String("pxc-anti-affinity-key", "kubernetes.io/hostname", "Pod anti-affinity rules. Allowed values: none, kubernetes.io/hostname, failure-domain.beta.kubernetes.io/zone, failure-domain.beta.kubernetes.io/region")

	createCmd.Flags().Int32("proxy-instances", 1, "Number of ProxySQL nodes in cluster")
	createCmd.Flags().String("proxy-request-cpu", "600m", "ProxySQL node requests for CPU, in cores. (500m = .5 cores)")
	createCmd.Flags().String("proxy-request-mem", "1G", "ProxySQL node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	createCmd.Flags().String("proxy-anti-affinity-key", "kubernetes.io/hostname", "Pod anti-affinity rules. Allowed values: none, kubernetes.io/hostname, failure-domain.beta.kubernetes.io/zone, failure-domain.beta.kubernetes.io/region")

	createCmd.Flags().String("s3-endpoint-url", "", "Endpoing URL of S3 compatible storage to store backup at")
	createCmd.Flags().String("s3-bucket", "", "Bucket of S3 compatible storage to store backup at")
	createCmd.Flags().String("s3-region", "", "Region of S3 compatible storage to store backup at")
	createCmd.Flags().String("s3-credentials-secret", "", "Secrets with credentials for S3 compatible storage to store backup at. Alternatevily you can set --s3-access-key-id and --s3-secret-access-key instead.")
	createCmd.Flags().String("s3-access-key-id", "", "Access Key ID for S3 compatible storage to store backup at")
	createCmd.Flags().String("s3-secret-access-key", "", "Access Key for S3 compatible storage to store backup at")
	createCmd.Flags().Bool("pmm-enabled", false, "Activate monitoring service.")
	createCmd.Flags().String("pmm-image", "perconalab/pmm-client:1.17.1", "Monitoring service client image.")
	createCmd.Flags().String("pmm-server-host", "monitoring-service", "Monitoring service server hostname.")
	createCmd.Flags().String("pmm-server-user", "", "Monitoring service user for login.")
	createCmd.Flags().String("pmm-server-password", "", "Monitoring service password for login.")
	skipS3Storage = createCmd.Flags().Bool("s3-skip-storage", false, "Don't create S3 compatible backup storage. Has to be set manually later on.")
	envCrt = createCmd.Flags().String("environment", "", "Target kubernetes cluster")
	labels = createCmd.Flags().String("labels", "", "PXC cluster labels inside kubernetes/openshift cluster")
	operatorImage = createCmd.Flags().String("operator-image", "", "Custom operator image")

	PXCCmd.AddCommand(createCmd)
}

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
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/engines/pxc/types/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	defaultVersion = "default"
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
		var instance dbaas.Instance
		dbaas.CreateDB(instance)
		/*
			labelsMap := make(map[string]string)
			if len(*labels) > 0 {
				keyValues := strings.Split(*labels, ",")
				for index := range keyValues {
					itemSlice := strings.Split(keyValues[index], "=")
					labelsMap[itemSlice[0]] = itemSlice[1]
				}
			}

			pxcOperator, err := pxc.NewPXCController(labelsMap, *envCrt)
			if err != nil {
				log.Error("new pxc operator: ", err)
				return
			}

			config, err := parseCreateFlagsToConfig(cmd.Flags(), args[0])
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
			err = pxcOperator.CreateDBCluster(config, *operatorImage)
			if err != nil {
				log.Error(errors.Wrap(err, "create DB"))
				return
			}

			time.Sleep(1 * time.Minute)
			getStatusMaxTries := 1200
			tries := 0
			tckr := time.NewTicker(500 * time.Millisecond)
			defer tckr.Stop()
			for range tckr.C {
				cluster, err := pxcOperator.CheckDBClusterStatus(args[0])
				if err != nil {
					log.Error("check status: ", err)
					return
				}
				switch cluster.State {
				case pxc.AppStateReady:
					sp.Stop()
					log.WithField("data", cluster).Info("Cluster ready")
					return
				}

				if tries >= getStatusMaxTries {
					log.Error("unable to start cluster")
					return
				}
				tries++
			}

			sp.Stop()
		*/
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
	skipS3Storage = createCmd.Flags().Bool("s3-skip-storage", true, "Don't create S3 compatible backup storage. Has to be set manually later on.")
	envCrt = createCmd.Flags().String("environment", "", "Target kubernetes cluster")
	labels = createCmd.Flags().String("labels", "", "PXC cluster labels inside kubernetes/openshift cluster")
	operatorImage = createCmd.Flags().String("operator-image", "", "Custom operator image")

	PXCCmd.AddCommand(createCmd)
}

func parseCreateFlagsToConfig(f *pflag.FlagSet, ClusterName string) (config config.ClusterConfig, err error) {
	config.Name = ClusterName
	config.PXC.StorageSize, err = f.GetString("storage-size")
	if err != nil {
		return config, errors.New("undefined `storage size`")
	}
	config.PXC.StorageClass, err = f.GetString("storage-class")
	if err != nil {
		return config, errors.New("undefined `storage class`")
	}
	config.PXC.Instances, err = f.GetInt32("pxc-instances")
	if err != nil {
		return config, errors.New("undefined `pxc-instances`")
	}
	config.PXC.RequestCPU, err = f.GetString("pxc-request-cpu")
	if err != nil {
		return config, errors.New("undefined `pxc-request-cpu`")
	}
	config.PXC.RequestMem, err = f.GetString("pxc-request-mem")
	if err != nil {
		return config, errors.New("undefined `pxc-request-mem`")
	}
	config.PXC.AntiAffinityKey, err = f.GetString("pxc-anti-affinity-key")
	if err != nil {
		return config, errors.New("undefined `pxc-anti-affinity-key`")
	}

	config.ProxySQL.Instances, err = f.GetInt32("proxy-instances")
	if err != nil {
		return config, errors.New("undefined `proxy-instances`")
	}
	if config.ProxySQL.Instances > 0 {
		config.ProxySQL.RequestCPU, err = f.GetString("proxy-request-cpu")
		if err != nil {
			return config, errors.New("undefined `proxy-request-cpu`")
		}
		config.ProxySQL.RequestMem, err = f.GetString("proxy-request-mem")
		if err != nil {
			return config, errors.New("undefined `proxy-request-mem`")
		}
		config.ProxySQL.AntiAffinityKey, err = f.GetString("proxy-anti-affinity-key")
		if err != nil {
			return config, errors.New("undefined `proxy-anti-affinity-key`")
		}
	}

	skipS3Storage, err := f.GetBool("s3-skip-storage")
	if err != nil {
		return config, errors.New("undefined `s3-skip-storage`")
	}

	if !skipS3Storage {
		config.S3.EndpointURL, err = f.GetString("s3-endpoint-url")
		if err != nil {
			return config, errors.New("undefined `s3-endpoint-url`")
		}
		config.S3.Bucket, err = f.GetString("s3-bucket")
		if err != nil {
			return config, errors.New("undefined `s3-bucket`")
		}
		config.S3.Region, err = f.GetString("s3-region")
		if err != nil {
			return config, errors.New("undefined `s3-region`")
		}
		config.S3.CredentialsSecret, err = f.GetString("s3-credentials-secret")
		if err != nil {
			return config, errors.New("undefined `s3-credentials-secret`")
		}
		config.S3.KeyID, err = f.GetString("s3-access-key-id")
		if err != nil {
			return config, errors.New("undefined `s3-access-key-id`")
		}
		config.S3.Key, err = f.GetString("s3-secret-access-key")
		if err != nil {
			return config, errors.New("undefined `s3-secret-access-key`")
		}
	}

	config.PMM.Enabled, err = f.GetBool("pmm-enabled")
	if err != nil {
		return config, errors.New("undefined `pmm-enabled`")
	}

	if config.PMM.Enabled {
		config.PMM.Image, err = f.GetString("pmm-image")
		if err != nil {
			return config, errors.New("undefined `pmm-image`")
		}
		config.PMM.ServerHost, err = f.GetString("pmm-server-host")
		if err != nil {
			return config, errors.New("undefined `pmm-server-host`")
		}
		config.PMM.ServerUser, err = f.GetString("pmm-server-user")
		if err != nil {
			return config, errors.New("undefined `pmm-server-user`")
		}
		config.PMM.ServerPass, err = f.GetString("pmm-server-password")
		if err != nil {
			return config, errors.New("undefined `pmm-server-password`")
		}
	}

	return
}

package command

import (
	"log"

	"github.com/Percona-Lab/percona-dbaas-cli/cmd/percona-dbaas/service-broker/server"
	"github.com/spf13/cobra"
)

// PxcBrokerCmd represents the pxc broker command
var PxcBrokerCmd = &cobra.Command{
	Use:   "pxc-broker",
	Short: "Start PXC broker",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("starting broker")
		server, err := server.New("8081", cmd.Flags())
		if err != nil {
			log.Println(err)
			return
		}
		server.Start()
	},
}
var skipS3Storage *bool

func init() {
	PxcBrokerCmd.Flags().String("storage-size", "6G", "PXC node volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)")
	PxcBrokerCmd.Flags().String("storage-class", "", "Name of the StorageClass required by the volume claim")
	PxcBrokerCmd.Flags().Int32("pxc-instances", 3, "Number of PXC nodes in cluster")
	PxcBrokerCmd.Flags().String("pxc-request-cpu", "600m", "PXC node requests for CPU, in cores. (500m = .5 cores)")
	PxcBrokerCmd.Flags().String("pxc-request-mem", "1G", "PXC node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	PxcBrokerCmd.Flags().String("pxc-anti-affinity-key", "kubernetes.io/hostname", "Pod anti-affinity rules. Allowed values: none, kubernetes.io/hostname, failure-domain.beta.kubernetes.io/zone, failure-domain.beta.kubernetes.io/region")

	PxcBrokerCmd.Flags().Int32("proxy-instances", 1, "Number of ProxySQL nodes in cluster")
	PxcBrokerCmd.Flags().String("proxy-request-cpu", "600m", "ProxySQL node requests for CPU, in cores. (500m = .5 cores)")
	PxcBrokerCmd.Flags().String("proxy-request-mem", "1G", "ProxySQL node requests for memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)")
	PxcBrokerCmd.Flags().String("proxy-anti-affinity-key", "kubernetes.io/hostname", "Pod anti-affinity rules. Allowed values: none, kubernetes.io/hostname, failure-domain.beta.kubernetes.io/zone, failure-domain.beta.kubernetes.io/region")

	PxcBrokerCmd.Flags().String("s3-endpoint-url", "", "Endpoing URL of S3 compatible storage to store backup at")
	PxcBrokerCmd.Flags().String("s3-bucket", "", "Bucket of S3 compatible storage to store backup at")
	PxcBrokerCmd.Flags().String("s3-region", "", "Region of S3 compatible storage to store backup at")
	PxcBrokerCmd.Flags().String("s3-credentials-secret", "", "Secrets with credentials for S3 compatible storage to store backup at. Alternatevily you can set --s3-key-id and --s3-key instead.")
	PxcBrokerCmd.Flags().String("s3-key-id", "", "Access Key ID for S3 compatible storage to store backup at")
	PxcBrokerCmd.Flags().String("s3-key", "", "Access Key for S3 compatible storage to store backup at")
	skipS3Storage = PxcBrokerCmd.Flags().Bool("s3-skip-storage", true, "Don't create S3 compatible backup storage. Has to be set manually later on.")
}

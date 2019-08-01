package servicebroker

import (
	"log"

	"github.com/Percona-Lab/percona-dbaas-cli/broker/server"
	"github.com/spf13/cobra"
)

// PxcBrokerCmd represents the pxc broker command
var PxcBrokerCmd = &cobra.Command{
	Use:   "service-broker",
	Short: "Start service broker",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Starting broker")
		server, err := server.NewBroker(cmd.Flag("port").Value.String())
		if err != nil {
			log.Println(err)
			return
		}
		server.Start()
	},
}

var skipS3Storage *bool

func init() {
	PxcBrokerCmd.Flags().String("port", "8081", "Broker API port")
}

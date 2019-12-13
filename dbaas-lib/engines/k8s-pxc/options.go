package pxc

import (
	"reflect"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/config"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/options"
)

// ParseOptions parse PXC options given in "object.paramValue=val,objectTwo.paramValue=val" string
func (p *PXC) ParseOptions(opts string) error {
	var c config.ClusterConfig

	res := config.PodResources{
		Requests: config.ResourcesList{
			CPU:    "600m",
			Memory: "1G",
		},
	}
	topologyKey := "none" //TODO: Deside what value is default "none" or "kubernetes.io/hostname"
	aff := config.PodAffinity{
		TopologyKey: topologyKey,
	}
	c.PXC.Size = int32(3)
	c.PXC.Resources = res
	c.PXC.Affinity = aff
	c.ProxySQL.Size = int32(1)
	c.ProxySQL.Resources = res
	c.ProxySQL.Affinity = aff
	c.S3.SkipStorage = true

	err := options.Parse(&c, reflect.TypeOf(c), opts)
	if err != nil {
		return err
	}
	p.config = c

	return nil
}

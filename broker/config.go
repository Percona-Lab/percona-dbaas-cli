package broker

import (
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

const (
	pxcServiceID     = "pxc-service-broker-id"
	pxcServiceName   = "percona-xtradb-cluster"
	psmdbServiseID   = "percona-server-for-mongodb"
	psmdbServiceName = "percona-server-for-mongodb"
)

func SetDefault(c *dbaas.ClusterConfig) {
	c.PXC.StorageSize = "6G"
	c.PXC.StorageClass = ""
	c.PXC.Instances = int32(3)
	c.PXC.RequestCPU = "600m"
	c.PXC.RequestMem = "1G"
	c.PXC.AntiAffinityKey = "kubernetes.io/hostname"
	c.ProxySQL.Instances = int32(1)
	c.ProxySQL.RequestCPU = "600m"
	c.ProxySQL.RequestMem = "1G"
	c.ProxySQL.AntiAffinityKey = "kubernetes.io/hostname"
	c.S3.SkipStorage = true
}

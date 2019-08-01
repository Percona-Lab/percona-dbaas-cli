package broker

import (
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
)

const (
	pxcServiceID     = "percona-xtradb-cluster-id"
	pxcServiceName   = "percona-xtradb-cluster"
	psmdbServiceID   = "percona-server-for-mongodb-id"
	psmdbServiceName = "percona-server-for-mongodb"
)

func SetPXCDefaults(c *dbaas.ClusterConfig) {
	c.PXC.StorageSize = "6G"
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

func SetPSMDBDefaults(c *dbaas.ClusterConfig) {
	c.PSMDB.StorageSize = "6G"
	c.PSMDB.Instances = int32(3)
	c.PSMDB.RequestCPU = "600m"
	c.PSMDB.RequestMem = "1G"
	c.PSMDB.AntiAffinityKey = "kubernetes.io/hostname"

	c.S3.SkipStorage = true
}

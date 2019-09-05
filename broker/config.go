package broker

import (
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/psmdb"
	"github.com/Percona-Lab/percona-dbaas-cli/dbaas/pxc"
)

const (
	pxcServiceID     = "percona-xtradb-cluster-id"
	pxcServiceName   = "percona-xtradb-cluster"
	psmdbServiceID   = "percona-server-for-mongodb-id"
	psmdbServiceName = "percona-server-for-mongodb"
)

func SetPXCDefaults(c *pxc.ClusterConfig) {
	c.PXC.StorageSize = "6G"
	c.PXC.Instances = int32(3)
	c.PXC.RequestCPU = "600m"
	c.PXC.RequestMem = "1G"
	c.PXC.AntiAffinityKey = "kubernetes.io/hostname"

	c.ProxySQL.Instances = int32(1)
	c.ProxySQL.RequestCPU = "600m"
	c.ProxySQL.RequestMem = "1G"
	c.ProxySQL.AntiAffinityKey = "kubernetes.io/hostname"

	c.PMM.Image = "perconalab/pmm-client:1.17.1"

	c.S3.SkipStorage = true
}

func SetPSMDBDefaults(c *psmdb.ClusterConfig) {
	c.PSMDB.StorageSize = "6G"
	c.PSMDB.Instances = int32(3)
	c.PSMDB.RequestCPU = "600m"
	c.PSMDB.RequestMem = "1G"
	c.PSMDB.AntiAffinityKey = "kubernetes.io/hostname"

	c.S3.SkipStorage = true
}

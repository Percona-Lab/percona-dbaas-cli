package pxc

type Config struct {
	StorageSize          string
	StorageClass         string
	Instances            int32
	RequestCPU           string
	RequestMem           string
	AntiAffinityKey      string
	ProxyInstances       int32
	ProxyRequestCPU      string
	ProxyRequestMem      string
	ProxyAntiAffinityKey string
	S3EndpointURL        string
	S3Bucket             string
	S3Region             string
	S3CredentialsSecret  string
	S3KeyID              string
	S3Key                string
	S3SkipStorage        bool
}

func (c *Config) SetDefault() {
	c.StorageSize = "6G"
	c.StorageClass = ""
	c.Instances = int32(3)
	c.RequestCPU = "600m"
	c.RequestMem = "1G"
	c.AntiAffinityKey = "kubernetes.io/hostname"
	c.ProxyInstances = int32(1)
	c.ProxyRequestCPU = "600m"
	c.ProxyRequestMem = "1G"
	c.ProxyAntiAffinityKey = "kubernetes.io/hostname"
	c.S3SkipStorage = true
}

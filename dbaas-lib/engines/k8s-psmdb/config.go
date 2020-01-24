package psmdb

type PSMDBCluster interface {
	Upgrade(imgs map[string]string)
	GetName() string
	SetName(name string)
	SetUsersSecretName(name string)
	MarshalRequests() error
	GetCR() (string, error)
	SetLabels(labels map[string]string)
	GetOperatorImage() string
	SetDefaults() error
}

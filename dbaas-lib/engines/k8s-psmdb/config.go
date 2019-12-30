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

type AppState string

const (
	AppStateUnknown AppState = "unknown"
	AppStateInit             = "initializing"
	AppStateReady            = "ready"
	AppStateError            = "error"
)

type PSMDBClusterStatus struct {
	Messages []string `json:"message,omitempty"`
	Status   AppState `json:"state,omitempty"`
}

type AppStatus struct {
	Size    int32    `json:"size,omitempty"`
	Ready   int32    `json:"ready,omitempty"`
	Status  AppState `json:"status,omitempty"`
	Message string   `json:"message,omitempty"`
}

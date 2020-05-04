package main

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

type k8sStatus struct {
	Status PSMDBClusterStatus
}

func GetK8SObject(typ, name string) ([]byte, error) {
	o, err := runCmd("kubectl", "get", typ, name, "-o", "json")
	if err != nil {
		return nil, err
	}
	return []byte(o), nil
}

func DeleteDeployment(name string) (string, error) {
	o, err := runCmd("kubectl", "delete", "deployment", name)
	if err != nil {
		return "", err
	}
	return o, nil
}

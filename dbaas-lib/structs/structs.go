package structs

import "fmt"

type DB struct {
	ResourceName     string `json:"resourceName,omitempty"`
	ResourceEndpoint string `json:"resourceEndpoint,omitempty"`
	Size             string `json:"size,omitempty"`
	Port             int    `json:"port,omitempty"`
	User             string `json:"user,omitempty"`
	Pass             string `json:"pass,omitempty"`
	Status           string `json:"status,omitempty"`
	Engine           string `json:"engine,omitempty"`
	Provider         string `json:"provider,omitempty"`
	Message          string `json:"message,omitempty"`
}

func (d DB) String() string {
	provider := ""
	if len(d.Provider) > 0 {
		provider = fmt.Sprintf("Provider:          %s", d.Provider)
	}
	engine := ""
	if len(d.Engine) > 0 {
		engine = fmt.Sprintf("\nEngine:            %s", d.Engine)
	}
	resourceName := ""
	if len(d.ResourceName) > 0 {
		resourceName = fmt.Sprintf("\nResource Name:     %s", d.ResourceName)
	}
	resourceEndpoint := ""
	if len(d.ResourceEndpoint) > 0 {
		resourceEndpoint = fmt.Sprintf("\nResource Endpoint: %s", d.ResourceEndpoint)
	}
	port := ""
	if d.Port > 0 {
		port = fmt.Sprintf("\nPort:              %d", d.Port)
	}
	user := ""
	if len(d.User) > 0 {
		user = fmt.Sprintf("\nUser:              %s", d.User)
	}
	pass := ""
	if len(d.Pass) > 0 {
		pass = fmt.Sprintf("\nPass:              %s", d.Pass)
	}
	message := ""
	if len(d.Message) > 0 {
		message = fmt.Sprintf("\n\n%s\n", d.Message)
	}

	return provider + engine + resourceName + resourceEndpoint + port + user + pass + message
}

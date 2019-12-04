package structs

import "fmt"

type DB struct {
	Name     string `json:"name,omitempty"`
	Size     string `json:"size,omitempty"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Pass     string `json:"pass,omitempty"`
	Status   string `json:"status,omitempty"`
	Engine   string `json:"engine,omitempty"`
	Provider string `json:"provider,omitempty"`
}

func (d DB) String() string {
	stringMsg := "Host: %s\nPort: %d\nUser: %s\nPass: %s"
	return fmt.Sprintf(stringMsg, d.Host, d.Port, d.User, d.Pass)
}

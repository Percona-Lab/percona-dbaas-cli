package main

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

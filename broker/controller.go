package broker

import "net/http"

//Controller represent handlers interface for broker API
type Controller interface {
	Catalog(w http.ResponseWriter, r *http.Request)
	CreateServiceInstance(w http.ResponseWriter, r *http.Request)
	GetServiceInstance(w http.ResponseWriter, r *http.Request)
	RemoveServiceInstance(w http.ResponseWriter, r *http.Request)
	Bind(w http.ResponseWriter, r *http.Request)
	UnBind(w http.ResponseWriter, r *http.Request)
}

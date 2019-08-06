package server

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Percona-Lab/percona-dbaas-cli/broker"
)

// Server is for handling broker API
type Server struct {
	controller broker.Controller
	port       string
}

// NewBroker return server for PXC broker
func NewBroker(port string) (*Server, error) {
	controller, err := broker.New()
	if err != nil {
		return nil, err
	}

	return &Server{
		controller: controller,
		port:       port,
	}, nil
}

// Start start handling requests
func (s *Server) Start() {
	router := mux.NewRouter()

	router.HandleFunc("/v2/catalog", s.controller.Catalog).Methods("GET")
	router.HandleFunc("/v2/service_instances", s.controller.GetServiceInstances).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.controller.GetServiceInstance).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/last_operation", s.controller.GetServiceInstanceLastOperation).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.controller.CreateServiceInstance).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.controller.UpdateServiceInstance).Methods("UPDATE")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.controller.RemoveServiceInstance).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.controller.Bind).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.controller.UnBind).Methods("DELETE")

	http.Handle("/", router)

	log.Println("Broker started, listening on port " + s.port + "...")
	http.ListenAndServe(":"+s.port, nil)
}

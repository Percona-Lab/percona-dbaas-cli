package server

import (
	"log"
	"net/http"
	"os"

	"github.com/Percona-Lab/percona-dbaas-cli/cmd/percona-dbaas/service-broker/pxc"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
)

type Server struct {
	controller *pxc.Controller
	port       string
}

func New(port string, flags *pflag.FlagSet) (*Server, error) {
	controller, err := pxc.New(flags)
	if err != nil {
		return nil, err
	}

	return &Server{
		controller: &controller,
		port:       port,
	}, nil
}

func (s *Server) Start() {
	router := mux.NewRouter()

	router.HandleFunc("/v2/catalog", s.controller.Catalog).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.controller.GetServiceInstance).Methods("GET")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.controller.CreateServiceInstance).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}", s.controller.RemoveServiceInstance).Methods("DELETE")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.controller.Bind).Methods("PUT")
	router.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", s.controller.UnBind).Methods("DELETE")

	http.Handle("/", router)

	cfPort := os.Getenv("PORT")
	if cfPort != "" {
		s.port = cfPort
	}

	log.Println("Broker started, listening on port " + s.port + "...")
	http.ListenAndServe(":"+s.port, nil)
}

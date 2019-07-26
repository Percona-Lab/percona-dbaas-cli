package pxc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/alecthomas/jsonschema"
	"github.com/gorilla/mux"
	"github.com/spf13/pflag"

	"github.com/Percona-Lab/percona-dbaas-cli/broker"
)

// Controller represents controller for broker API
type Controller struct {
	instanceMap map[string]*broker.ServiceInstance
	bindingMap  map[string]*broker.ServiceBinding
	flags       *pflag.FlagSet
}

// ProvisionParameters represents the parameters that can be tuned on a cluster
type ProvisionParameters struct {
	// ClusterName of the cluster resource
	ClusterName string

	// Replicas represents the number of nodes for cluster
	Replicas *int32

	//TopologyyKey
	TopologyKey string

	// Size represents the size. Example: 1Gi
	Size string
}

// BindParameters is for binding services
type BindParameters struct {
	// ClusterName - the database name
	ClusterName string

	// User - the username
	User string
}

// New creates new controller
func New(flags *pflag.FlagSet) (Controller, error) {
	var pxc Controller
	pxc.instanceMap = make(map[string]*broker.ServiceInstance)
	pxc.bindingMap = make(map[string]*broker.ServiceBinding)
	pxc.flags = flags

	return pxc, nil
}

func (c *Controller) Catalog(w http.ResponseWriter, r *http.Request) {
	log.Println("Get Service Broker Catalog...")

	planList := []broker.ServicePlan{
		broker.ServicePlan{
			Name:        "Default",
			ID:          "Default",
			Description: "",
			Metadata: &broker.ServicePlanMetadata{
				DisplayName: "Default",
			},
			Schemas: &broker.ServiceSchemas{
				Instance: broker.ServiceInstanceSchema{
					Create: mustGetJSONSchema(&ProvisionParameters{}),
				},
				Binding: broker.ServiceBindingSchema{
					Create: mustGetJSONSchema(&BindParameters{}),
				},
			},
			Bindable: true,
			Free:     true,
		},
	}

	var catalog = broker.Catalog{
		Services: []broker.Service{
			broker.Service{
				ID:          "pxc-service-broker-id",
				Name:        "percona-xtradb-cluster",
				Description: "database",
				Bindable:    true,
				Plans:       planList,
				Metadata: &broker.ServiceMetadata{
					DisplayName:         "Percona XtraDB Cluster Operator",
					LongDescription:     "Percona is Cloud Native",
					DocumentationURL:    "https://github.com/percona/percona-xtradb-cluster-operator",
					SupportURL:          "",
					ImageURL:            "https://www.percona.com/blog/wp-content/uploads/2016/06/Percona-XtraDB-Cluster-certification-1-300x250.png",
					ProviderDisplayName: "percona",
				},
				Tags: []string{
					"pxc",
				},
				PlanUpdateable: true,
			},
		},
	}

	WriteResponse(w, http.StatusOK, catalog)
}

const (
	defaultPolling = 10
)

func (c *Controller) CreateServiceInstance(w http.ResponseWriter, r *http.Request) {
	var params ProvisionParameters
	log.Println("Create Service Instance...")

	var instance broker.ServiceInstance

	err := ProvisionDataFromRequest(r, &instance)
	if err != nil {
		log.Println("Provision instatnce:", err)
	}

	p := instance.Parameters.(map[string]interface{})

	log.Println("Deploy cluster")
	if p["ClusterName"] != nil {
		params.ClusterName = p["ClusterName"].(string)
	}

	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")

	skipS3 := true
	err = c.DeployPXCCluster(params, &skipS3, instanceID)
	if err != nil {
		log.Println("Deploy cluster", err)
	}

	instance.InternalID = instanceID
	instance.DashboardURL = "http://dashbaord_url"
	instance.ID = ExtractVarsFromRequest(r, "service_instance_guid")
	instance.LastOperation = &broker.LastOperation{
		State:                    broker.InProgressOperationSate,
		Description:              broker.InProgressOperationDescription,
		AsyncPollIntervalSeconds: defaultPolling,
	}

	c.instanceMap[instance.ID] = &instance

	response := broker.CreateServiceInstanceResponse{
		DashboardURL:  instance.DashboardURL,
		LastOperation: instance.LastOperation,
	}
	WriteResponse(w, http.StatusAccepted, response)
}

func (c *Controller) GetServiceInstance(w http.ResponseWriter, r *http.Request) {
	log.Println("Get Service Instance State....")

	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := broker.CreateServiceInstanceResponse{
		DashboardURL:  instance.DashboardURL,
		LastOperation: instance.LastOperation,
	}
	WriteResponse(w, http.StatusOK, response)
}

func (c *Controller) RemoveServiceInstance(w http.ResponseWriter, r *http.Request) {
	log.Println("Remove Service Instance...")

	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		return
	}
	c.DeletePXCCluster("some-name")
	delete(c.instanceMap, instanceID)

	WriteResponse(w, http.StatusOK, "{}")
}

func (c *Controller) Bind(w http.ResponseWriter, r *http.Request) {
	log.Println("Bind Service Instance...")

	bindingID := ExtractVarsFromRequest(r, "service_binding_guid")
	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")

	instance := c.instanceMap[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	c.bindingMap[bindingID] = &broker.ServiceBinding{
		ID:                bindingID,
		ServiceID:         instance.ServiceID,
		ServicePlanID:     instance.PlanID,
		ServiceInstanceID: instance.ID,
	}

	credentials := broker.Credential{
		UserName:   "PXCUser",
		PublicIP:   "ServiceAddress",
		PrivateKey: "UserPass",
	}
	response := broker.CreateServiceBindingResponse{
		Credentials: credentials,
	}

	WriteResponse(w, http.StatusCreated, response)
}

func (c *Controller) UnBind(w http.ResponseWriter, r *http.Request) {
	log.Println("Unbind Service Instance...")

	bindingID := ExtractVarsFromRequest(r, "service_binding_guid")
	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		return
	}

	delete(c.bindingMap, bindingID)

	WriteResponse(w, http.StatusOK, "{}")
}

func (c *Controller) deleteAssociatedBindings(instanceID string) error {
	for id, binding := range c.bindingMap {
		if binding.ServiceInstanceID == instanceID {
			delete(c.bindingMap, id)
		}
	}

	return nil
}

func ProvisionDataFromRequest(r *http.Request, object interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, object)
	if err != nil {
		return err
	}

	return nil
}

func WriteResponse(w http.ResponseWriter, code int, object interface{}) {
	data, err := json.Marshal(object)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("X-Broker-API-Version", "2.13")
	w.WriteHeader(code)
	fmt.Fprintf(w, string(data))
}

func ExtractVarsFromRequest(r *http.Request, varName string) string {
	return mux.Vars(r)[varName]
}

// mustGetJSONSchema takes an struct{} and returns the related JSON schema
func mustGetJSONSchema(obj interface{}) broker.Schema {
	var reflector = jsonschema.Reflector{
		ExpandedStruct: true,
	}
	var schemaBytes, err = json.Marshal(reflector.Reflect(obj))
	if err != nil {
		panic(err)
	}
	schema := broker.Schema{}
	err = json.Unmarshal(schemaBytes, &schema.Parameters)
	if err != nil {
		panic(err)
	}

	return schema
}

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

	"github.com/Percona-Lab/percona-dbaas-cli/pxcbroker"
)

// Controller represents controller for broker API
type Controller struct {
	instanceMap map[string]*pxcbroker.ServiceInstance
	bindingMap  map[string]*pxcbroker.ServiceBinding
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

type BindParameters struct {
	// ClusterName - the database name
	ClusterName string

	// User - the username
	User string
}

// New creates new controller
func New(flags *pflag.FlagSet) (Controller, error) {
	var pxc Controller
	pxc.instanceMap = make(map[string]*pxcbroker.ServiceInstance)
	pxc.bindingMap = make(map[string]*pxcbroker.ServiceBinding)
	pxc.flags = flags

	return pxc, nil
}

func (c *Controller) Catalog(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Service Broker Catalog...")

	planList := []pxcbroker.ServicePlan{
		pxcbroker.ServicePlan{
			Name:        "Default",
			ID:          "Default",
			Description: "",
			Metadata: &pxcbroker.ServicePlanMetadata{
				DisplayName: "Default",
			},
			Schemas: &pxcbroker.ServiceSchemas{
				Instance: pxcbroker.ServiceInstanceSchema{
					Create: mustGetJSONSchema(&ProvisionParameters{}),
				},
				Binding: pxcbroker.ServiceBindingSchema{
					Create: mustGetJSONSchema(&BindParameters{}),
				},
			},
			Bindable: true,
			Free:     true,
		},
	}

	var catalog = pxcbroker.Catalog{
		Services: []pxcbroker.Service{
			pxcbroker.Service{
				ID:          "pxc-service-broker-id",
				Name:        "percona-xtradb-cluster",
				Description: "database",
				Bindable:    true,
				Plans:       planList,
				Metadata: &pxcbroker.ServiceMetadata{
					DisplayName:         "Percona XtraDB Cluster Operator",
					LongDescription:     "Percona is Cloud Native",
					DocumentationUrl:    "https://github.com/percona/percona-xtradb-cluster-operator",
					SupportUrl:          "",
					ImageUrl:            "https://www.percona.com/blog/wp-content/uploads/2016/06/Percona-XtraDB-Cluster-certification-1-300x250.png",
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

	var instance pxcbroker.ServiceInstance

	err := ProvisionDataFromRequest(r, &instance)
	if err != nil {
		log.Println("Provision instatnce:", err)
	}

	p := instance.Parameters.(map[string]interface{})

	log.Println("Deploy cluster")
	if p["ClusterName"] != nil {
		params.ClusterName = p["ClusterName"].(string)
	}
	skipS3 := true
	err = c.DeployPXCCluster(params, &skipS3)
	if err != nil {
		log.Println("Deploy cluster", err)
	}

	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")

	instance.InternalId = instanceID
	instance.DashboardUrl = "http://dashbaord_url"
	instance.Id = ExtractVarsFromRequest(r, "service_instance_guid")
	instance.LastOperation = &pxcbroker.LastOperation{
		State:                    "in progress",
		Description:              "creating service instance...",
		AsyncPollIntervalSeconds: defaultPolling,
	}

	c.instanceMap[instance.Id] = &instance

	response := pxcbroker.CreateServiceInstanceResponse{
		DashboardUrl:  instance.DashboardUrl,
		LastOperation: instance.LastOperation,
	}
	WriteResponse(w, http.StatusAccepted, response)
}

func (c *Controller) GetServiceInstance(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get Service Instance State....")

	instanceId := ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceId]
	if instance == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	state := "running"
	if state == "pending" {
		instance.LastOperation.State = "in progress"
		instance.LastOperation.Description = "creating service instance..."
	} else if state == "running" {
		instance.LastOperation.State = "succeeded"
		instance.LastOperation.Description = "successfully created service instance"
	} else {
		instance.LastOperation.State = "failed"
		instance.LastOperation.Description = "failed to create service instance"
	}

	response := pxcbroker.CreateServiceInstanceResponse{
		DashboardUrl:  instance.DashboardUrl,
		LastOperation: instance.LastOperation,
	}
	WriteResponse(w, http.StatusOK, response)
}

func (c *Controller) RemoveServiceInstance(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Remove Service Instance...")

	instanceId := ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceId]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		return
	}
	c.DeletePXCCluster("some-name")
	delete(c.instanceMap, instanceId)

	WriteResponse(w, http.StatusOK, "{}")
}

func (c *Controller) Bind(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Bind Service Instance...")
	/*
		bindingId := ExtractVarsFromRequest(r, "service_binding_guid")
		instanceId := ExtractVarsFromRequest(r, "service_instance_guid")

		instance := c.instanceMap[instanceId]
		if instance == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		c.bindingMap[bindingId] = &broker.ServiceBinding{
			Id:                bindingId,
			ServiceId:         instance.ServiceId,
			ServicePlanId:     instance.PlanId,
			PrivateKey:        privateKey,
			ServiceInstanceId: instance.Id,
		}
	*/
	response := pxcbroker.CreateServiceBindingResponse{}
	WriteResponse(w, http.StatusCreated, response)
}

func (c *Controller) UnBind(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Unbind Service Instance...")

	/*	bindingId := utils.ExtractVarsFromRequest(r, "service_binding_guid")
		instanceId := utils.ExtractVarsFromRequest(r, "service_instance_guid")
		instance := c.instanceMap[instanceId]
		if instance == nil {
			w.WriteHeader(http.StatusGone)
			return
		}

		err := c.cloudClient.RevokeKeyPair(instance.InternalId, c.bindingMap[bindingId].PrivateKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		delete(c.bindingMap, bindingId)
	*/
	WriteResponse(w, http.StatusOK, "{}")
}

func (c *Controller) deleteAssociatedBindings(instanceId string) error {
	for id, binding := range c.bindingMap {
		if binding.ServiceInstanceId == instanceId {
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
func mustGetJSONSchema(obj interface{}) pxcbroker.Schema {
	var reflector = jsonschema.Reflector{
		ExpandedStruct: true,
	}
	var schemaBytes, err = json.Marshal(reflector.Reflect(obj))
	if err != nil {
		panic(err)
	}
	schema := pxcbroker.Schema{}
	err = json.Unmarshal(schemaBytes, &schema.Parameters)
	if err != nil {
		panic(err)
	}

	return schema
}

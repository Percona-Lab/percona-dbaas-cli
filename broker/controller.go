package broker

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas"
	"github.com/Percona-Lab/percona-dbaas-cli/gcloud"
	"github.com/alecthomas/jsonschema"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Controller represents controller for broker API
type Controller struct {
	instanceMap map[string]*ServiceInstance
	bindingMap  map[string]*ServiceBinding
	dbaas       *dbaas.Cmd
	EnvName     string
}

// ProvisionParameters represents the parameters that can be tuned on a cluster
type PXCProvisionParameters struct {
	// ClusterName of the cluster resource
	ClusterName string `json:"cluster_name"`

	// Replicas represents the number of nodes for cluster
	Replicas int32 `json:"replicas,omitempty"`

	//TopologyyKey
	TopologyKey string `json:"topology_key,omitempty"`

	// Size represents the size. Example: 1Gi
	Size string `json:"size,omitempty"`
}

// ProvisionParameters represents the parameters that can be tuned on a cluster
type PSMDBProvisionParameters struct {
	// ClusterName of the cluster resource
	ClusterName string `json:"cluster_name"`

	// Replicas represents the number of nodes for cluster
	Replicas int32 `json:"replicas,omitempty"`

	//TopologyyKey
	TopologyKey string `json:"topology_key,omitempty"`

	// Size represents the size. Example: 1Gi
	Size string `json:"size,omitempty"`
}

// BindParameters is for binding services
type BindParameters struct {
	// ClusterName - the database name
	ClusterName string

	// User - the username
	User string
}

// New creates new controller
func New() (Controller, error) {
	var c Controller
	dbservice, err := dbaas.New("")
	if err != nil {
		return c, errors.Wrap(err, "create dbaas")
	}
	c.dbaas = dbservice
	c.instanceMap = make(map[string]*ServiceInstance)
	c.bindingMap = make(map[string]*ServiceBinding)
	err = c.getBrokerInstances("pxc")
	if err != nil {
		log.Println(errors.Wrap(err, "get pxc instances"))
	}
	err = c.getBrokerInstances("psmdb")
	if err != nil {
		log.Println(errors.Wrap(err, "get psmdb instances"))
	}

	return c, nil
}

type GCloudRequest struct {
	EnvName   string `json:"envName"`
	Project   string `json:"project"`
	Zone      string `json:"zone"`
	Cluster   string `json:"cluster"`
	KeyFile   string `json:"keyFile"`
	Namespace string `json:"namespace"`
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}
func (c *Controller) Gcloud(w http.ResponseWriter, r *http.Request) {
	var gc GCloudRequest
	err := ProvisionDataFromRequest(r, &gc)
	if err != nil {
		log.Println("Provision data:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(gc.KeyFile) > 0 {
		if isJSON(gc.KeyFile) {
			gc.KeyFile = base64.StdEncoding.EncodeToString([]byte(gc.KeyFile))
		} else {
			dat, err := ioutil.ReadFile(gc.KeyFile)
			if err != nil {
				fmt.Printf("\n[error] %s\n", err)
				return
			}
			gc.KeyFile = base64.StdEncoding.EncodeToString([]byte(dat))
		}
	}
	cloudEnv, err := gcloud.New(gc.EnvName, gc.Project, gc.Zone, gc.Cluster, gc.KeyFile, gc.Namespace)
	if err != nil {
		fmt.Printf("\n[error] %s\n", err)
		return
	}
	err = cloudEnv.Setup()
	if err != nil {
		fmt.Printf("\n[error] %s\n", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type EnvironmentRequest struct {
	Environment string `json:"environment"`
}

func (c *Controller) Environment(w http.ResponseWriter, r *http.Request) {
	var envReq EnvironmentRequest
	err := ProvisionDataFromRequest(r, &envReq)
	if err != nil {
		log.Println("Provision data:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c.EnvName = envReq.Environment
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) Catalog(w http.ResponseWriter, r *http.Request) {
	log.Println("Get Service Broker Catalog...")

	PXCPlanList := []ServicePlan{
		{
			Name:        "percona-xtradb-cluster",
			ID:          "percona-xtradb-id",
			Description: "percona xtradb cluster",
			Metadata: &ServicePlanMetadata{
				DisplayName: "standard",
			},
			Schemas: &ServiceSchemas{
				Instance: ServiceInstanceSchema{
					Create: mustGetJSONSchema(&PXCProvisionParameters{}),
				},
				Binding: ServiceBindingSchema{
					Create: mustGetJSONSchema(&BindParameters{}),
				},
			},
			Bindable: true,
			Free:     true,
		},
	}

	PSMDBPlanList := []ServicePlan{
		{
			Name:        "percona-server-for-mongodb",
			ID:          "percona-server-for-mongodb-id",
			Description: "percona server for mongodbr",
			Metadata: &ServicePlanMetadata{
				DisplayName: "standard",
			},
			Schemas: &ServiceSchemas{
				Instance: ServiceInstanceSchema{
					Create: mustGetJSONSchema(&PSMDBProvisionParameters{}),
				},
				Binding: ServiceBindingSchema{
					Create: mustGetJSONSchema(&BindParameters{}),
				},
			},
			Bindable: true,
			Free:     true,
		},
	}

	var catalog = Catalog{
		Services: []Service{
			{
				ID:          pxcServiceID,
				Name:        pxcServiceName,
				Description: "database",
				Bindable:    true,
				Plans:       PXCPlanList,
				Metadata: &ServiceMetadata{
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
			{
				ID:          psmdbServiceID,
				Name:        psmdbServiceName,
				Description: "database",
				Bindable:    true,
				Plans:       PSMDBPlanList,
				Metadata: &ServiceMetadata{
					DisplayName:         "Percona Kubernetes Operator for Percona Server for MongoDB",
					LongDescription:     "Percona is Cloud Native",
					DocumentationURL:    "https://www.percona.com/doc/kubernetes-operator-for-psmongodb/index.html",
					SupportURL:          "",
					ImageURL:            "https://www.percona.com/blog/wp-content/uploads/2016/04/Percona_ServerfMDBLogoVert.png",
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

func (c *Controller) GetServiceInstances(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, http.StatusOK, c.instanceMap)
}

const (
	defaultPolling = 10
)

func (c *Controller) CreateServiceInstance(w http.ResponseWriter, r *http.Request) {
	log.Println("Create Service Instance...")

	var instance ServiceInstance

	err := ProvisionDataFromRequest(r, &instance)
	if err != nil {
		log.Println("Provision instatnce:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")

	skipS3 := true

	instance.InternalID = instanceID
	instance.ID = ExtractVarsFromRequest(r, "service_instance_guid")
	c.DeployCluster(instance, &skipS3, instanceID)
	instance.LastOperation = &LastOperation{
		State:                    InProgressOperationSate,
		Description:              InProgressOperationDescription,
		AsyncPollIntervalSeconds: defaultPolling,
	}

	c.instanceMap[instance.ID] = &instance

	response := CreateServiceInstanceResponse{
		LastOperation: instance.LastOperation,
	}

	WriteResponse(w, http.StatusAccepted, response)
}

func (c *Controller) UpdateServiceInstance(w http.ResponseWriter, r *http.Request) {
	log.Println("Update Service Instance...")

	var instance ServiceInstance

	err := ProvisionDataFromRequest(r, &instance)
	if err != nil {
		log.Println("Provision instatnce:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")

	instance.InternalID = instanceID
	instance.ID = ExtractVarsFromRequest(r, "service_instance_guid")
	instance.LastOperation = &LastOperation{
		State:                    InProgressOperationSate,
		Description:              InProgressOperationDescription,
		AsyncPollIntervalSeconds: defaultPolling,
	}

	err = c.UpdateCluster(&instance)
	if err != nil {
		log.Println("Update instatnce:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	c.instanceMap[instance.ID] = &instance

	response := CreateServiceInstanceResponse{
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

	response := GetServiceInstanceLastOperationResponse{
		instance.LastOperation,
	}

	WriteResponse(w, http.StatusOK, response)
}

func (c *Controller) getBrokerInstances(typ string) error {
	s, err := c.dbaas.GetServiceBrokerInstances(typ)
	if err != nil {
		return errors.Wrap(err, "getBrokerInstances")
	}

	s = s[1 : len(s)-1]

	instances := bytes.Split(s, []byte("} {"))
	for k, v := range instances {
		var b ServiceInstance
		switch k {
		case 0:
			if len(instances) > 1 {
				v = append(v, []byte("}")...)
			}
			err = json.Unmarshal(v, &b)
			if err != nil {
				return errors.Wrap(err, "instance unmarshal")
			}
		case len(instances) - 1:
			v = append([]byte("{"), v...)
			err = json.Unmarshal(v, &b)
			if err != nil {
				return errors.Wrap(err, "instance unmarshal")
			}
		default:
			v = append([]byte("{"), s...)
			v = append(v, []byte("}")...)
			err = json.Unmarshal(v, &b)
			if err != nil {
				return errors.Wrap(err, "instance unmarshal")
			}
		}
		c.instanceMap[b.ID] = &b
	}
	return nil
}

func (c *Controller) GetServiceInstanceLastOperation(w http.ResponseWriter, r *http.Request) {
	log.Println("Get Service Instance Last Operaton...")

	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")
	instance := c.instanceMap[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := CreateServiceInstanceResponse{
		LastOperation: instance.LastOperation,
	}
	WriteResponse(w, http.StatusOK, response)
}

func (c *Controller) RemoveServiceInstance(w http.ResponseWriter, r *http.Request) {
	log.Println("Remove Service Instance...")
	for k, v := range c.instanceMap {
		log.Println(k)
		log.Println(v)
	}
	instanceID := ExtractVarsFromRequest(r, "service_instance_guid")
	log.Println(instanceID)
	instance := c.instanceMap[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		return
	}

	err := c.DeleteCluster(instance)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	c.bindingMap[bindingID] = &ServiceBinding{
		ID:                bindingID,
		ServiceID:         instance.ServiceID,
		ServicePlanID:     instance.PlanID,
		ServiceInstanceID: instance.ID,
	}

	credentials := Credential{
		UserName:   "User",
		PublicIP:   "ServiceAddress",
		PrivateKey: "UserPass",
	}
	response := CreateServiceBindingResponse{
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

	w.WriteHeader(code)
	fmt.Fprintf(w, string(data))
}

func ExtractVarsFromRequest(r *http.Request, varName string) string {
	return mux.Vars(r)[varName]
}

// mustGetJSONSchema takes an struct{} and returns the related JSON schema
func mustGetJSONSchema(obj interface{}) Schema {
	var reflector = jsonschema.Reflector{
		ExpandedStruct: true,
	}
	var schemaBytes, err = json.Marshal(reflector.Reflect(obj))
	if err != nil {
		panic(err)
	}
	schema := Schema{}
	err = json.Unmarshal(schemaBytes, &schema.Parameters)
	if err != nil {
		panic(err)
	}

	return schema
}

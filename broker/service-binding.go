package broker

type ServiceBinding struct {
	ID                string `json:"id"`
	ServiceID         string `json:"service_id"`
	AppID             string `json:"app_id"`
	ServicePlanID     string `json:"service_plan_id"`
	PrivateKey        string `json:"private_key"`
	ServiceInstanceID string `json:"service_instance_id"`
}

type CreateServiceBindingResponse struct {
	Credentials interface{} `json:"credentials"`
}

type Credential struct {
	PublicIP   string `json:"public_ip"`
	UserName   string `json:"username"`
	PrivateKey string `json:"private_key"`
}

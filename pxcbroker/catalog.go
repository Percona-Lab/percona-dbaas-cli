package pxcbroker

type Service struct {
	Name           string   `json:"name"`
	ID             string   `json:"id"`
	Description    string   `json:"description"`
	Bindable       bool     `json:"bindable"`
	PlanUpdateable bool     `json:"plan_updateable, omitempty"`
	Tags           []string `json:"tags, omitempty"`
	Requires       []string `json:"requires, omitempty"`

	Metadata        interface{}   `json:"metadata, omitempty"`
	Plans           []ServicePlan `json:"plans"`
	DashboardClient interface{}   `json:"dashboard_client"`
}

type ServiceMetadata struct {
	DisplayName         string `json:"displayName,omitempty"`
	ImageUrl            string `json:"imageUrl,omitempty"`
	LongDescription     string `json:"longDescription,omitempty"`
	ProviderDisplayName string `json:"providerDisplayName,omitempty"`
	DocumentationUrl    string `json:"documentationUrl,omitempty"`
	SupportUrl          string `json:"supportUrl,omitempty"`
	Shareable           *bool  `json:"shareable,omitempty"`
	AdditionalMetadata  map[string]interface{}
}

type ServicePlan struct {
	Name        string          `json:"name"`
	ID          string          `json:"id"`
	Description string          `json:"description"`
	Metadata    interface{}     `json:"metadata, omitempty"`
	Schemas     *ServiceSchemas `json:"schemas,omitempty"`
	Bindable    bool            `json:"bindable"`
	Free        bool            `json:"free, omitempty"`
}

type ServicePlanMetadata struct {
	DisplayName        string   `json:"displayName,omitempty"`
	Bullets            []string `json:"bullets,omitempty"`
	AdditionalMetadata map[string]interface{}
}

type Catalog struct {
	Services []Service `json:"services"`
}

type ServiceSchemas struct {
	Instance ServiceInstanceSchema `json:"service_instance,omitempty"`
	Binding  ServiceBindingSchema  `json:"service_binding,omitempty"`
}

type ServiceInstanceSchema struct {
	Create Schema `json:"create,omitempty"`
	Update Schema `json:"update,omitempty"`
}

type ServiceBindingSchema struct {
	Create Schema `json:"create,omitempty"`
}

type Schema struct {
	Parameters map[string]interface{} `json:"parameters"`
}

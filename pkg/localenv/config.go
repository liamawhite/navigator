package localenv

// ServiceSpec defines how to create a service in the local environment
type ServiceSpec struct {
	// Name is the name of the service
	Name string

	// Replicas is the number of replicas for the service (ignored for external services)
	Replicas int32

	// Type determines the type of service to create
	Type ServiceType

	// ExternalIPs is used for external services to define manual endpoints
	ExternalIPs []string

	// NextService is used for microservice topology chaining - this service will call the next one
	NextService string
}

// ServiceType represents different types of services that can be deployed
type ServiceType string

const (
	// ServiceTypeWeb creates a regular ClusterIP service with deployment
	ServiceTypeWeb ServiceType = "web"

	// ServiceTypeHeadless creates a headless service (ClusterIP: None)
	ServiceTypeHeadless ServiceType = "headless"

	// ServiceTypeExternal creates a service with manual endpoints (no deployment)
	ServiceTypeExternal ServiceType = "external"

	// ServiceTypeTopology creates a service that can proxy to other services in a chain
	ServiceTypeTopology ServiceType = "topology"
)

// Scenario defines a collection of services that represent a specific demo scenario
type Scenario struct {
	// Name is the human-readable name of the scenario
	Name string

	// Description provides details about what this scenario demonstrates
	Description string

	// Services is the list of services to deploy for this scenario
	Services []ServiceSpec

	// IstioEnabled determines if this scenario should enable Istio injection
	IstioEnabled bool

	// Namespaces is a list of additional namespaces to create (beyond the default)
	Namespaces []string
}

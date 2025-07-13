package localenv

// Predefined scenarios for local development and demos

var (
	// ScenarioBasic provides a simple single web service for basic testing
	ScenarioBasic = &Scenario{
		Name:        "basic",
		Description: "A single web service for basic Navigator functionality testing",
		Services: []ServiceSpec{
			{
				Name:     "web-service",
				Replicas: 2,
				Type:     ServiceTypeWeb,
			},
		},
		IstioEnabled: false,
	}

	// ScenarioMicroserviceTopology creates a chain of services that call each other
	ScenarioMicroserviceTopology = &Scenario{
		Name:        "microservice-topology",
		Description: "A chain of three microservices demonstrating service-to-service communication",
		Services: []ServiceSpec{
			{
				Name:        "service-a",
				Replicas:    1,
				Type:        ServiceTypeTopology,
				NextService: "service-b",
			},
			{
				Name:        "service-b",
				Replicas:    1,
				Type:        ServiceTypeTopology,
				NextService: "service-c",
			},
			{
				Name:     "service-c",
				Replicas: 1,
				Type:     ServiceTypeTopology,
			},
		},
		IstioEnabled: false,
	}

	// ScenarioMixed demonstrates different types of Kubernetes services
	ScenarioMixed = &Scenario{
		Name:        "mixed-services",
		Description: "A mix of web, headless, and external services showcasing different service types",
		Services: []ServiceSpec{
			{
				Name:     "web-service",
				Replicas: 2,
				Type:     ServiceTypeWeb,
			},
			{
				Name:     "headless-service",
				Replicas: 1,
				Type:     ServiceTypeHeadless,
			},
			{
				Name:        "external-service",
				Type:        ServiceTypeExternal,
				ExternalIPs: []string{"203.0.113.1", "203.0.113.2"},
			},
		},
		IstioEnabled: false,
	}

	// ScenarioIstioDemo showcases Istio service mesh capabilities
	ScenarioIstioDemo = &Scenario{
		Name:        "istio-demo",
		Description: "Services with Istio sidecars for service mesh demonstration",
		Services: []ServiceSpec{
			{
				Name:        "frontend",
				Replicas:    2,
				Type:        ServiceTypeTopology,
				NextService: "backend",
			},
			{
				Name:        "backend",
				Replicas:    1,
				Type:        ServiceTypeTopology,
				NextService: "database",
			},
			{
				Name:     "database",
				Replicas: 1,
				Type:     ServiceTypeWeb,
			},
		},
		IstioEnabled: true,
	}

	// ScenarioHighScale demonstrates Navigator's performance with many services
	ScenarioHighScale = &Scenario{
		Name:        "high-scale",
		Description: "Multiple services with higher replica counts for performance testing",
		Services: []ServiceSpec{
			{
				Name:     "web-1",
				Replicas: 3,
				Type:     ServiceTypeWeb,
			},
			{
				Name:     "web-2",
				Replicas: 3,
				Type:     ServiceTypeWeb,
			},
			{
				Name:     "web-3",
				Replicas: 2,
				Type:     ServiceTypeWeb,
			},
			{
				Name:     "api-service",
				Replicas: 2,
				Type:     ServiceTypeWeb,
			},
			{
				Name:        "external-db",
				Type:        ServiceTypeExternal,
				ExternalIPs: []string{"203.0.113.10"},
			},
		},
		IstioEnabled: false,
	}

	// ScenarioComplexTopology creates a more complex microservice architecture
	ScenarioComplexTopology = &Scenario{
		Name:        "complex-topology",
		Description: "A complex microservice architecture with multiple service chains",
		Services: []ServiceSpec{
			{
				Name:        "api-gateway",
				Replicas:    2,
				Type:        ServiceTypeTopology,
				NextService: "user-service",
			},
			{
				Name:        "user-service",
				Replicas:    2,
				Type:        ServiceTypeTopology,
				NextService: "auth-service",
			},
			{
				Name:     "auth-service",
				Replicas: 1,
				Type:     ServiceTypeWeb,
			},
			{
				Name:        "order-service",
				Replicas:    2,
				Type:        ServiceTypeTopology,
				NextService: "payment-service",
			},
			{
				Name:     "payment-service",
				Replicas: 1,
				Type:     ServiceTypeWeb,
			},
			{
				Name:        "notification-service",
				Type:        ServiceTypeExternal,
				ExternalIPs: []string{"203.0.113.20"},
			},
		},
		IstioEnabled: false,
	}
)

// GetScenarioByName returns a scenario by its name
func GetScenarioByName(name string) *Scenario {
	scenarios := map[string]*Scenario{
		"basic":                 ScenarioBasic,
		"microservice-topology": ScenarioMicroserviceTopology,
		"mixed-services":        ScenarioMixed,
		"istio-demo":            ScenarioIstioDemo,
		"high-scale":            ScenarioHighScale,
		"complex-topology":      ScenarioComplexTopology,
	}

	return scenarios[name]
}

// ListScenarios returns all available scenarios
func ListScenarios() []*Scenario {
	return []*Scenario{
		ScenarioBasic,
		ScenarioMicroserviceTopology,
		ScenarioMixed,
		ScenarioIstioDemo,
		ScenarioHighScale,
		ScenarioComplexTopology,
	}
}

// GetScenarioNames returns the names of all available scenarios
func GetScenarioNames() []string {
	scenarios := ListScenarios()
	names := make([]string, len(scenarios))
	for i, scenario := range scenarios {
		names[i] = scenario.Name
	}
	return names
}

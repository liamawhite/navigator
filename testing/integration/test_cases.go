package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// TestCase represents a single integration test scenario
type TestCase struct {
	Name       string
	Setup      []ServiceSpec
	Timeout    time.Duration
	Assertions []Assertion
	ShouldFail bool
	SkipReason string
}

// Assertion represents a test assertion to be performed
type Assertion struct {
	Type     AssertionType
	Target   AssertionTarget
	Expected interface{}
	Message  string
}

// AssertionType defines the type of assertion
type AssertionType string

const (
	AssertionTypeEqual        AssertionType = "equal"
	AssertionTypeNotEqual     AssertionType = "not_equal"
	AssertionTypeGreater      AssertionType = "greater"
	AssertionTypeGreaterEqual AssertionType = "greater_equal"
	AssertionTypeLess         AssertionType = "less"
	AssertionTypeLessEqual    AssertionType = "less_equal"
	AssertionTypeLen          AssertionType = "len"
	AssertionTypeNotEmpty     AssertionType = "not_empty"
	AssertionTypeEmpty        AssertionType = "empty"
	AssertionTypeContains     AssertionType = "contains"
	AssertionTypeError        AssertionType = "error"
	AssertionTypeNoError      AssertionType = "no_error"
	AssertionTypeTrue         AssertionType = "true"
	AssertionTypeFalse        AssertionType = "false"
	AssertionTypeRegex        AssertionType = "regex"
)

// AssertionTarget defines what to assert against
type AssertionTarget struct {
	Type      TargetType
	Namespace string
	ServiceID string
	Field     string
}

// TargetType defines the type of target for assertions
type TargetType string

const (
	TargetTypeListServices  TargetType = "list_services"
	TargetTypeGetService    TargetType = "get_service"
	TargetTypeServiceCount  TargetType = "service_count"
	TargetTypeServiceName   TargetType = "service_name"
	TargetTypeServiceID     TargetType = "service_id"
	TargetTypeInstanceCount TargetType = "instance_count"
	TargetTypeInstanceIP    TargetType = "instance_ip"
	TargetTypeInstancePod   TargetType = "instance_pod"
	TargetTypeError         TargetType = "error"
)

// GetTestCases returns all integration test cases
func GetTestCases() []TestCase {
	return []TestCase{
		{
			Name: "basic_service_discovery",
			Setup: []ServiceSpec{
				{
					Name:     "test-service",
					Replicas: 1,
					Type:     ServiceTypeWeb,
				},
			},
			Timeout: 2 * time.Minute,
			Assertions: []Assertion{
				{
					Type:     AssertionTypeLen,
					Target:   AssertionTarget{Type: TargetTypeListServices},
					Expected: 1,
					Message:  "Should discover exactly one service",
				},
				{
					Type:     AssertionTypeEqual,
					Target:   AssertionTarget{Type: TargetTypeServiceName, Field: "test-service"},
					Expected: "test-service",
					Message:  "Service name should match",
				},
				{
					Type:     AssertionTypeGreater,
					Target:   AssertionTarget{Type: TargetTypeInstanceCount, ServiceID: "test-service"},
					Expected: 0,
					Message:  "Service should have at least one instance",
				},
			},
		},
		{
			Name: "multiple_services",
			Setup: []ServiceSpec{
				{Name: "web-1", Replicas: 1, Type: ServiceTypeWeb},
				{Name: "web-2", Replicas: 1, Type: ServiceTypeWeb},
				{Name: "api-service", Replicas: 1, Type: ServiceTypeWeb},
			},
			Timeout: 3 * time.Minute,
			Assertions: []Assertion{
				{
					Type:     AssertionTypeLen,
					Target:   AssertionTarget{Type: TargetTypeListServices},
					Expected: 3,
					Message:  "Should discover all three services",
				},
				{
					Type:    AssertionTypeNoError,
					Target:  AssertionTarget{Type: TargetTypeGetService, ServiceID: "web-1"},
					Message: "Should be able to get web-1 service",
				},
				{
					Type:    AssertionTypeNoError,
					Target:  AssertionTarget{Type: TargetTypeGetService, ServiceID: "web-2"},
					Message: "Should be able to get web-2 service",
				},
				{
					Type:    AssertionTypeNoError,
					Target:  AssertionTarget{Type: TargetTypeGetService, ServiceID: "api-service"},
					Message: "Should be able to get api-service",
				},
			},
		},
		{
			Name: "microservice_topology",
			Setup: []ServiceSpec{
				{Name: "service-a", Replicas: 1, Type: ServiceTypeTopology, NextService: "service-b"},
				{Name: "service-b", Replicas: 1, Type: ServiceTypeTopology, NextService: "service-c"},
				{Name: "service-c", Replicas: 1, Type: ServiceTypeTopology},
			},
			Timeout: 5 * time.Minute,
			Assertions: []Assertion{
				{
					Type:     AssertionTypeLen,
					Target:   AssertionTarget{Type: TargetTypeListServices},
					Expected: 3,
					Message:  "Should discover all services in topology",
				},
				{
					Type:     AssertionTypeEqual,
					Target:   AssertionTarget{Type: TargetTypeInstanceCount, ServiceID: "service-a"},
					Expected: 1,
					Message:  "service-a should have exactly one instance",
				},
				{
					Type:     AssertionTypeEqual,
					Target:   AssertionTarget{Type: TargetTypeInstanceCount, ServiceID: "service-b"},
					Expected: 1,
					Message:  "service-b should have exactly one instance",
				},
				{
					Type:     AssertionTypeEqual,
					Target:   AssertionTarget{Type: TargetTypeInstanceCount, ServiceID: "service-c"},
					Expected: 1,
					Message:  "service-c should have exactly one instance",
				},
				{
					Type:     AssertionTypeRegex,
					Target:   AssertionTarget{Type: TargetTypeInstanceIP, ServiceID: "service-a"},
					Expected: `^10\.`,
					Message:  "Instance IP should be in cluster network range",
				},
				{
					Type:     AssertionTypeContains,
					Target:   AssertionTarget{Type: TargetTypeInstancePod, ServiceID: "service-a"},
					Expected: "service-a",
					Message:  "Pod name should contain service name",
				},
			},
		},
		{
			Name: "mixed_service_types",
			Setup: []ServiceSpec{
				{Name: "web-service", Replicas: 2, Type: ServiceTypeWeb},
				{Name: "headless-service", Replicas: 1, Type: ServiceTypeHeadless},
				{Name: "external-service", Type: ServiceTypeExternal, ExternalIPs: []string{"203.0.113.1", "203.0.113.2"}},
			},
			Timeout: 4 * time.Minute,
			Assertions: []Assertion{
				{
					Type:     AssertionTypeLen,
					Target:   AssertionTarget{Type: TargetTypeListServices},
					Expected: 3,
					Message:  "Should discover all service types",
				},
				{
					Type:     AssertionTypeGreater,
					Target:   AssertionTarget{Type: TargetTypeInstanceCount, ServiceID: "web-service"},
					Expected: 0,
					Message:  "Web service should have instances",
				},
				{
					Type:     AssertionTypeEqual,
					Target:   AssertionTarget{Type: TargetTypeInstanceCount, ServiceID: "external-service"},
					Expected: 2,
					Message:  "External service should have 2 instances",
				},
			},
		},
		{
			Name: "error_handling",
			Setup: []ServiceSpec{
				{Name: "test-service", Replicas: 1, Type: ServiceTypeWeb},
			},
			Timeout: 2 * time.Minute,
			Assertions: []Assertion{
				{
					Type:    AssertionTypeError,
					Target:  AssertionTarget{Type: TargetTypeGetService, ServiceID: "nonexistent"},
					Message: "Should fail to get nonexistent service",
				},
				{
					Type:     AssertionTypeLen,
					Target:   AssertionTarget{Type: TargetTypeListServices, Namespace: "empty-namespace"},
					Expected: 0,
					Message:  "Empty namespace should have no services",
				},
			},
		},
		{
			Name: "namespace_isolation",
			Setup: []ServiceSpec{
				{Name: "isolated-service", Replicas: 1, Type: ServiceTypeWeb},
			},
			Timeout: 3 * time.Minute,
			Assertions: []Assertion{
				{
					Type:     AssertionTypeLen,
					Target:   AssertionTarget{Type: TargetTypeListServices},
					Expected: 1,
					Message:  "Should have one service in test namespace",
				},
				{
					Type:     AssertionTypeGreater,
					Target:   AssertionTarget{Type: TargetTypeListServices, Namespace: "all"},
					Expected: 1,
					Message:  "Should have services from multiple namespaces when listing all",
				},
			},
		},
		{
			Name: "high_replica_count",
			Setup: []ServiceSpec{
				{Name: "scaled-service", Replicas: 3, Type: ServiceTypeWeb},
			},
			Timeout: 4 * time.Minute,
			Assertions: []Assertion{
				{
					Type:     AssertionTypeEqual,
					Target:   AssertionTarget{Type: TargetTypeInstanceCount, ServiceID: "scaled-service"},
					Expected: 3,
					Message:  "Should have all 3 replicas as instances",
				},
			},
		},
		{
			Name: "external_endpoints_validation",
			Setup: []ServiceSpec{
				{Name: "external-validation", Type: ServiceTypeExternal, ExternalIPs: []string{"203.0.113.10"}},
			},
			Timeout: 2 * time.Minute,
			Assertions: []Assertion{
				{
					Type:     AssertionTypeEqual,
					Target:   AssertionTarget{Type: TargetTypeInstanceCount, ServiceID: "external-validation"},
					Expected: 1,
					Message:  "Should have exactly one external instance",
				},
				{
					Type:     AssertionTypeEqual,
					Target:   AssertionTarget{Type: TargetTypeInstanceIP, ServiceID: "external-validation"},
					Expected: "203.0.113.10",
					Message:  "Should have correct external IP",
				},
				{
					Type:    AssertionTypeEmpty,
					Target:  AssertionTarget{Type: TargetTypeInstancePod, ServiceID: "external-validation"},
					Message: "External service should not have pod references",
				},
			},
		},
	}
}

// RunTestCase executes a single test case against the given environment
func RunTestCase(t *testing.T, env TestEnvironment, testCase TestCase) {
	if testCase.SkipReason != "" {
		t.Skip(testCase.SkipReason)
	}

	t.Run(testCase.Name, func(t *testing.T) {
		require.NoError(t, env.Setup(t))
		defer env.Cleanup(t)

		ctx, cancel := context.WithTimeout(context.Background(), testCase.Timeout)
		defer cancel()

		// Create services
		if len(testCase.Setup) > 0 {
			require.NoError(t, env.CreateServices(ctx, testCase.Setup))

			// Wait for web services to be ready
			var serviceNames []string
			for _, spec := range testCase.Setup {
				if spec.Type == ServiceTypeWeb || spec.Type == ServiceTypeTopology {
					serviceNames = append(serviceNames, spec.Name)
				}
			}
			if len(serviceNames) > 0 {
				require.NoError(t, env.WaitForServices(ctx, serviceNames))
			}
		}

		client := env.GetGRPCClient()
		namespace := env.GetNamespace()

		// Execute assertions
		for _, assertion := range testCase.Assertions {
			executeAssertion(t, ctx, client, namespace, assertion)
		}
	})
}

// executeAssertion performs a single assertion
func executeAssertion(t *testing.T, ctx context.Context, client v1alpha1.ServiceRegistryServiceClient, namespace string, assertion Assertion) {
	switch assertion.Target.Type {
	case TargetTypeListServices:
		targetNamespace := assertion.Target.Namespace
		if targetNamespace == "" {
			targetNamespace = namespace
		}

		var namespacePtr *string
		if targetNamespace != "" && targetNamespace != "all" {
			namespacePtr = &targetNamespace
		}
		// For "all" namespace, namespacePtr remains nil which lists all namespaces

		resp, err := client.ListServices(ctx, &v1alpha1.ListServicesRequest{
			Namespace: namespacePtr,
		})

		if assertion.Type == AssertionTypeError {
			assert.Error(t, err, assertion.Message)
			return
		}

		require.NoError(t, err, assertion.Message)

		// For length assertions, pass the slice directly
		if assertion.Type == AssertionTypeLen {
			performAssertion(t, assertion, resp.Services)
		} else {
			performAssertion(t, assertion, len(resp.Services))
		}

	case TargetTypeGetService:
		serviceID := namespace + "/" + assertion.Target.ServiceID
		resp, err := client.GetService(ctx, &v1alpha1.GetServiceRequest{
			Id: serviceID,
		})

		if assertion.Type == AssertionTypeError {
			assert.Error(t, err, assertion.Message)
			return
		}

		if assertion.Type == AssertionTypeNoError {
			assert.NoError(t, err, assertion.Message)
			return
		}

		require.NoError(t, err, assertion.Message)

		switch assertion.Target.Field {
		case "name":
			performAssertion(t, assertion, resp.Service.Name)
		case "id":
			performAssertion(t, assertion, resp.Service.Id)
		default:
			performAssertion(t, assertion, resp.Service)
		}

	case TargetTypeInstanceCount:
		serviceID := namespace + "/" + assertion.Target.ServiceID
		resp, err := client.GetService(ctx, &v1alpha1.GetServiceRequest{
			Id: serviceID,
		})
		require.NoError(t, err, assertion.Message)

		// For length assertions, pass the slice directly
		if assertion.Type == AssertionTypeLen {
			performAssertion(t, assertion, resp.Service.Instances)
		} else {
			performAssertion(t, assertion, len(resp.Service.Instances))
		}

	case TargetTypeInstanceIP:
		serviceID := namespace + "/" + assertion.Target.ServiceID
		resp, err := client.GetService(ctx, &v1alpha1.GetServiceRequest{
			Id: serviceID,
		})
		require.NoError(t, err, assertion.Message)
		require.Greater(t, len(resp.Service.Instances), 0, "Service should have at least one instance")
		performAssertion(t, assertion, resp.Service.Instances[0].Ip)

	case TargetTypeInstancePod:
		serviceID := namespace + "/" + assertion.Target.ServiceID
		resp, err := client.GetService(ctx, &v1alpha1.GetServiceRequest{
			Id: serviceID,
		})
		require.NoError(t, err, assertion.Message)
		require.Greater(t, len(resp.Service.Instances), 0, "Service should have at least one instance")
		performAssertion(t, assertion, resp.Service.Instances[0].Pod)

	case TargetTypeServiceName:
		resp, err := client.ListServices(ctx, &v1alpha1.ListServicesRequest{
			Namespace: &namespace,
		})
		require.NoError(t, err, assertion.Message)

		// Find service by name from the Field
		serviceName := assertion.Target.Field
		for _, service := range resp.Services {
			if service.Name == serviceName {
				performAssertion(t, assertion, service.Name)
				return
			}
		}

		t.Fatalf("Service with name '%s' not found in list", serviceName)

	default:
		t.Fatalf("Unknown target type: %s", assertion.Target.Type)
	}
}

// performAssertion executes the actual assertion
func performAssertion(t *testing.T, assertion Assertion, actual interface{}) {
	switch assertion.Type {
	case AssertionTypeEqual:
		assert.Equal(t, assertion.Expected, actual, assertion.Message)
	case AssertionTypeNotEqual:
		assert.NotEqual(t, assertion.Expected, actual, assertion.Message)
	case AssertionTypeGreater:
		assert.Greater(t, actual, assertion.Expected, assertion.Message)
	case AssertionTypeGreaterEqual:
		assert.GreaterOrEqual(t, actual, assertion.Expected, assertion.Message)
	case AssertionTypeLess:
		assert.Less(t, actual, assertion.Expected, assertion.Message)
	case AssertionTypeLessEqual:
		assert.LessOrEqual(t, actual, assertion.Expected, assertion.Message)
	case AssertionTypeLen:
		assert.Len(t, actual, assertion.Expected.(int), assertion.Message)
	case AssertionTypeNotEmpty:
		assert.NotEmpty(t, actual, assertion.Message)
	case AssertionTypeEmpty:
		assert.Empty(t, actual, assertion.Message)
	case AssertionTypeContains:
		assert.Contains(t, actual, assertion.Expected, assertion.Message)
	case AssertionTypeTrue:
		assert.True(t, actual.(bool), assertion.Message)
	case AssertionTypeFalse:
		assert.False(t, actual.(bool), assertion.Message)
	case AssertionTypeRegex:
		assert.Regexp(t, assertion.Expected, actual, assertion.Message)
	case AssertionTypeError:
		assert.Error(t, actual.(error), assertion.Message)
	case AssertionTypeNoError:
		assert.NoError(t, actual.(error), assertion.Message)
	default:
		t.Fatalf("Unknown assertion type: %s", assertion.Type)
	}
}

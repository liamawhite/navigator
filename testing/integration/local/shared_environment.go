package local

import (
	"context"
	"fmt"
	"testing"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/testing/integration"
)

// SharedEnvironment implements TestEnvironment using a shared Kind cluster
type SharedEnvironment struct {
	sharedCluster *SharedCluster
	namespace     string
	fixtures      *TestFixtures
}

// Setup implements TestEnvironment.Setup
func (e *SharedEnvironment) Setup(t *testing.T) error {
	// Create test namespace
	e.fixtures = NewTestFixtures(e.sharedCluster.client, e.namespace)

	ctx := context.Background()
	if err := e.fixtures.CreateNamespace(ctx); err != nil {
		return err
	}

	// Clean up namespace when test is done
	t.Cleanup(func() {
		if err := e.fixtures.DeleteNamespace(context.Background()); err != nil {
			t.Logf("Failed to clean up namespace %s: %v", e.namespace, err)
		}
	})

	return nil
}

// Cleanup implements TestEnvironment.Cleanup
func (e *SharedEnvironment) Cleanup(t *testing.T) error {
	// Cleanup is handled by t.Cleanup() function in Setup()
	return nil
}

// GetGRPCClient implements TestEnvironment.GetGRPCClient
func (e *SharedEnvironment) GetGRPCClient() v1alpha1.ServiceRegistryServiceClient {
	return e.sharedCluster.grpcClient
}

// GetNamespace implements TestEnvironment.GetNamespace
func (e *SharedEnvironment) GetNamespace() string {
	return e.namespace
}

// CreateServices implements TestEnvironment.CreateServices
func (e *SharedEnvironment) CreateServices(ctx context.Context, services []integration.ServiceSpec) error {
	for _, spec := range services {
		switch spec.Type {
		case integration.ServiceTypeWeb:
			if err := e.fixtures.CreateWebService(ctx, spec.Name, spec.Replicas); err != nil {
				return err
			}
		case integration.ServiceTypeHeadless:
			if err := e.fixtures.CreateHeadlessService(ctx, spec.Name); err != nil {
				return err
			}
		case integration.ServiceTypeExternal:
			if err := e.fixtures.CreateExternalService(ctx, spec.Name, spec.ExternalIPs); err != nil {
				return err
			}
		case integration.ServiceTypeTopology:
			if err := e.fixtures.CreateTopologyService(ctx, spec.Name, spec.Replicas, spec.NextService); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported service type: %s", spec.Type)
		}
	}
	return nil
}

// WaitForServices implements TestEnvironment.WaitForServices
func (e *SharedEnvironment) WaitForServices(ctx context.Context, serviceNames []string) error {
	for _, serviceName := range serviceNames {
		if err := e.fixtures.WaitForDeployment(ctx, serviceName); err != nil {
			return err
		}
		if err := e.fixtures.WaitForService(ctx, serviceName); err != nil {
			return err
		}
	}
	return nil
}

// DeleteServices implements TestEnvironment.DeleteServices
func (e *SharedEnvironment) DeleteServices(ctx context.Context, serviceNames []string) error {
	// For shared environment, we clean up the entire namespace
	// Individual service deletion can be implemented if needed
	return nil
}

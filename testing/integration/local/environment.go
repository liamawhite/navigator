package local

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	grpcserver "github.com/liamawhite/navigator/internal/grpc"
	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	kubeconfigds "github.com/liamawhite/navigator/pkg/datastore/kubeconfig"
	"github.com/liamawhite/navigator/testing/integration"
)

// LocalEnvironment implements TestEnvironment for local Kind clusters
type LocalEnvironment struct {
	cluster     *KindCluster
	client      kubernetes.Interface
	fixtures    *TestFixtures
	grpcServer  *grpcserver.Server
	grpcClient  v1alpha1.ServiceRegistryServiceClient
	grpcConn    *grpc.ClientConn
	kubeconfig  string
	namespace   string
	clusterName string
}

// NewLocalEnvironment creates a new local test environment
func NewLocalEnvironment(clusterName string) *LocalEnvironment {
	return &LocalEnvironment{
		clusterName: clusterName,
		namespace:   "navigator-integration-test",
	}
}

// Setup implements TestEnvironment.Setup
func (e *LocalEnvironment) Setup(t *testing.T) error {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Setup Kind cluster
	e.cluster = NewKindCluster(e.clusterName)
	t.Cleanup(func() {
		e.cluster.Cleanup()
		if e.cluster.Exists(context.Background()) {
			if err := e.cluster.Delete(context.Background()); err != nil {
				t.Logf("Failed to clean up cluster: %v", err)
			}
		}
	})

	// Clean up any existing cluster
	if e.cluster.Exists(ctx) {
		if err := e.cluster.Delete(ctx); err != nil {
			return err
		}
	}

	// Create new cluster
	if err := e.cluster.Create(ctx); err != nil {
		return err
	}

	// Get kubeconfig
	kubeconfig, err := e.cluster.GetKubeconfig(ctx)
	if err != nil {
		return err
	}
	e.kubeconfig = kubeconfig

	// Create Kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	e.client = client

	// Setup test namespace
	e.fixtures = NewTestFixtures(client, e.namespace)
	if err := e.fixtures.CreateNamespace(ctx); err != nil {
		return err
	}
	t.Cleanup(func() {
		if err := e.fixtures.DeleteNamespace(context.Background()); err != nil {
			t.Logf("Failed to clean up namespace: %v", err)
		}
	})

	// Create Navigator datastore
	datastore, err := kubeconfigds.New(e.kubeconfig)
	if err != nil {
		return err
	}

	// Start Navigator server
	e.grpcServer, err = grpcserver.NewServer(datastore, 0) // Use any available port
	if err != nil {
		return err
	}

	// Start server in goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- e.grpcServer.Start()
	}()

	// Give server time to start
	time.Sleep(2 * time.Second)

	t.Cleanup(func() {
		e.grpcServer.Stop()
		// Check if server exited with error
		select {
		case err := <-serverErrChan:
			if err != nil {
				t.Logf("Server stopped with error: %v", err)
			}
		default:
			// Server is still running or stopped gracefully
		}
	})

	// Connect to gRPC server
	conn, err := grpc.Dial(e.grpcServer.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	e.grpcConn = conn
	e.grpcClient = v1alpha1.NewServiceRegistryServiceClient(conn)

	t.Cleanup(func() {
		if e.grpcConn != nil {
			e.grpcConn.Close()
		}
	})

	return nil
}

// Cleanup implements TestEnvironment.Cleanup
func (e *LocalEnvironment) Cleanup(t *testing.T) error {
	// Cleanup is handled by t.Cleanup() functions in Setup()
	return nil
}

// GetGRPCClient implements TestEnvironment.GetGRPCClient
func (e *LocalEnvironment) GetGRPCClient() v1alpha1.ServiceRegistryServiceClient {
	return e.grpcClient
}

// GetNamespace implements TestEnvironment.GetNamespace
func (e *LocalEnvironment) GetNamespace() string {
	return e.namespace
}

// CreateServices implements TestEnvironment.CreateServices
func (e *LocalEnvironment) CreateServices(ctx context.Context, services []integration.ServiceSpec) error {
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
func (e *LocalEnvironment) WaitForServices(ctx context.Context, serviceNames []string) error {
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
func (e *LocalEnvironment) DeleteServices(ctx context.Context, serviceNames []string) error {
	// For local environment, we typically clean up the entire namespace
	// Individual service deletion can be implemented if needed
	return nil
}

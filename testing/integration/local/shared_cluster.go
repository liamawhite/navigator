package local

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	grpcserver "github.com/liamawhite/navigator/internal/grpc"
	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	kubeconfigds "github.com/liamawhite/navigator/pkg/datastore/kubeconfig"
)

// SharedCluster manages a single Kind cluster shared across multiple tests
type SharedCluster struct {
	cluster    *KindCluster
	client     kubernetes.Interface
	grpcServer *grpcserver.Server
	grpcConn   *grpc.ClientConn
	grpcClient v1alpha1.ServiceRegistryServiceClient
	kubeconfig string
	mu         sync.Mutex
}

// NewSharedCluster creates a new shared cluster
func NewSharedCluster(clusterName string) (*SharedCluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	sc := &SharedCluster{
		cluster: NewKindCluster(clusterName),
	}

	// Clean up any existing cluster
	if sc.cluster.Exists(ctx) {
		if err := sc.cluster.Delete(ctx); err != nil {
			return nil, err
		}
	}

	// Create new cluster
	if err := sc.cluster.Create(ctx); err != nil {
		return nil, err
	}

	// Get kubeconfig
	kubeconfig, err := sc.cluster.GetKubeconfig(ctx)
	if err != nil {
		return nil, err
	}
	sc.kubeconfig = kubeconfig

	// Create Kubernetes client
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	sc.client = client

	// Install Istio
	if err := sc.cluster.InstallIstio(ctx); err != nil {
		return nil, fmt.Errorf("failed to install Istio: %w", err)
	}

	// Enable Istio injection for default namespace
	if err := sc.cluster.EnableIstioInjection(ctx, "default"); err != nil {
		return nil, fmt.Errorf("failed to enable Istio injection: %w", err)
	}

	// Create Navigator datastore
	datastore, err := kubeconfigds.New(sc.kubeconfig)
	if err != nil {
		return nil, err
	}

	// Start Navigator server
	sc.grpcServer, err = grpcserver.NewServer(datastore, 0) // Use any available port
	if err != nil {
		return nil, err
	}

	// Start server in goroutine
	go func() {
		if err := sc.grpcServer.Start(); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Give server time to start
	time.Sleep(2 * time.Second)

	// Connect to gRPC server
	conn, err := grpc.NewClient(sc.grpcServer.Address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	sc.grpcConn = conn
	sc.grpcClient = v1alpha1.NewServiceRegistryServiceClient(conn)

	return sc, nil
}

// NewEnvironment creates a new test environment with a random namespace
func (sc *SharedCluster) NewEnvironment() *SharedEnvironment {
	namespace := generateRandomNamespace()
	return &SharedEnvironment{
		sharedCluster: sc,
		namespace:     namespace,
	}
}

// Cleanup cleans up the shared cluster
func (sc *SharedCluster) Cleanup() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.grpcConn != nil {
		_ = sc.grpcConn.Close()
	}

	if sc.grpcServer != nil {
		sc.grpcServer.Stop()
	}

	if sc.cluster != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if sc.cluster.Exists(ctx) {
			_ = sc.cluster.Delete(ctx)
		}
	}
}

// generateRandomNamespace generates a random namespace name
func generateRandomNamespace() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	const length = 8

	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result[i] = chars[num.Int64()]
	}

	return fmt.Sprintf("test-%s", string(result))
}

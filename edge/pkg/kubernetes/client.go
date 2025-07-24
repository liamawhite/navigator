// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubernetes

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	typesv1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client and provides service discovery functionality
type Client struct {
	clientset   kubernetes.Interface
	istioClient istioclient.Interface
	restConfig  *rest.Config
	logger      *slog.Logger
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string, logger *slog.Logger) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		// Use kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
		}
	} else {
		// Use in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	istioClient, err := istioclient.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create istio client: %w", err)
	}

	return &Client{
		clientset:   clientset,
		istioClient: istioClient,
		restConfig:  config,
		logger:      logger,
	}, nil
}

// GetClientset returns the underlying Kubernetes clientset
func (k *Client) GetClientset() kubernetes.Interface {
	return k.clientset
}

// GetRestConfig returns the underlying Kubernetes REST config
func (k *Client) GetRestConfig() *rest.Config {
	return k.restConfig
}

// GetClusterState discovers all services in the cluster and returns the cluster state
func (k *Client) GetClusterState(ctx context.Context) (*v1alpha1.ClusterState, error) {
	// Parallelize API calls and map building in single goroutines
	var wg sync.WaitGroup
	var servicesResult *corev1.ServiceList
	var endpointSlicesByService map[string][]discoveryv1.EndpointSlice
	var podsByName map[string]*corev1.Pod
	var protoDestinationRules []*typesv1alpha1.DestinationRule
	var protoEnvoyFilters []*typesv1alpha1.EnvoyFilter
	var protoRequestAuthentications []*typesv1alpha1.RequestAuthentication
	var protoGateways []*typesv1alpha1.Gateway
	var protoSidecars []*typesv1alpha1.Sidecar
	var protoVirtualServices []*typesv1alpha1.VirtualService
	var protoIstioControlPlaneConfig *typesv1alpha1.IstioControlPlaneConfig

	// Create error channel to collect errors from all goroutines
	errChan := make(chan error, 10)
	wg.Add(10)

	// Fetch Kubernetes resources concurrently
	go k.fetchServices(ctx, &wg, &servicesResult, errChan)
	go k.fetchEndpointSlices(ctx, &wg, &endpointSlicesByService, errChan)
	go k.fetchPods(ctx, &wg, &podsByName, errChan)

	// Fetch and convert Istio resources concurrently
	go k.fetchDestinationRules(ctx, &wg, &protoDestinationRules, errChan)
	go k.fetchEnvoyFilters(ctx, &wg, &protoEnvoyFilters, errChan)
	go k.fetchRequestAuthentications(ctx, &wg, &protoRequestAuthentications, errChan)
	go k.fetchGateways(ctx, &wg, &protoGateways, errChan)
	go k.fetchSidecars(ctx, &wg, &protoSidecars, errChan)
	go k.fetchVirtualServices(ctx, &wg, &protoVirtualServices, errChan)
	go k.fetchIstioControlPlaneConfig(ctx, &wg, &protoIstioControlPlaneConfig, errChan)

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect all errors from the channel
	var errors []error
	for err := range errChan {
		if err != nil {
			errors = append(errors, err)
		}
	}

	// If we have any errors, merge them and return
	if len(errors) > 0 {
		return nil, k.mergeErrors(errors)
	}

	// Convert services using the fetched data
	var protoServices []*v1alpha1.Service
	for _, svc := range servicesResult.Items {
		protoService := k.convertServiceWithMaps(&svc, endpointSlicesByService, podsByName)
		protoServices = append(protoServices, protoService)
	}

	return &v1alpha1.ClusterState{
		Services:                 protoServices,
		DestinationRules:         protoDestinationRules,
		EnvoyFilters:             protoEnvoyFilters,
		RequestAuthentications:   protoRequestAuthentications,
		Gateways:                 protoGateways,
		Sidecars:                 protoSidecars,
		VirtualServices:          protoVirtualServices,
		IstioControlPlaneConfig:  protoIstioControlPlaneConfig,
	}, nil
}

// mergeErrors combines multiple errors into a single error with detailed information
func (k *Client) mergeErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}
	if len(errors) == 1 {
		return errors[0]
	}

	var errorMessages []string
	for _, err := range errors {
		errorMessages = append(errorMessages, err.Error())
	}

	return fmt.Errorf("multiple errors occurred (%d total): %s",
		len(errors),
		strings.Join(errorMessages, "; "))
}

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
	"sync"

	istioclient "istio.io/client-go/pkg/clientset/versioned"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	istioextlisters "istio.io/client-go/pkg/listers/extensions/v1alpha1"
	istiov1alpha3listers "istio.io/client-go/pkg/listers/networking/v1alpha3"
	istiov1beta1listers "istio.io/client-go/pkg/listers/networking/v1beta1"
	istioseclisters "istio.io/client-go/pkg/listers/security/v1beta1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appsv1lister "k8s.io/client-go/listers/apps/v1"
	corev1lister "k8s.io/client-go/listers/core/v1"
	discoveryv1lister "k8s.io/client-go/listers/discovery/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client and provides service discovery functionality
type Client struct {
	clientset   kubernetes.Interface
	istioClient istioclient.Interface
	restConfig  *rest.Config
	logger      *slog.Logger

	mu      sync.Mutex
	started bool

	// informer factories (nil until Start is called)
	k8sFactory   informers.SharedInformerFactory
	istioFactory istioinformers.SharedInformerFactory

	// k8s listers (populated by Start)
	servicesLister       corev1lister.ServiceLister
	podsLister           corev1lister.PodLister
	endpointSlicesLister discoveryv1lister.EndpointSliceLister
	deploymentsLister    appsv1lister.DeploymentLister
	namespacesLister     corev1lister.NamespaceLister

	// Istio listers (populated by Start)
	destinationRulesLister       istiov1beta1listers.DestinationRuleLister
	gatewaysLister               istiov1beta1listers.GatewayLister
	sidecarsLister               istiov1beta1listers.SidecarLister
	virtualServicesLister        istiov1beta1listers.VirtualServiceLister
	serviceEntriesLister         istiov1beta1listers.ServiceEntryLister
	envoyFiltersLister           istiov1alpha3listers.EnvoyFilterLister
	requestAuthenticationsLister istioseclisters.RequestAuthenticationLister
	peerAuthenticationsLister    istioseclisters.PeerAuthenticationLister
	authorizationPoliciesLister  istioseclisters.AuthorizationPolicyLister
	wasmPluginsLister            istioextlisters.WasmPluginLister
}

// NewClient creates a new Kubernetes client
func NewClient(kubeconfigPath string, logger *slog.Logger) (*Client, error) {
	return NewClientWithContext(kubeconfigPath, "", logger)
}

// NewClientWithContext creates a new Kubernetes client with a specific context
func NewClientWithContext(kubeconfigPath string, contextName string, logger *slog.Logger) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfigPath != "" {
		// Build config with specific context if provided
		if contextName != "" {
			// Load kubeconfig and override context
			kubeconfig, err := clientcmd.LoadFromFile(kubeconfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
			}

			// Override the current context
			overrides := &clientcmd.ConfigOverrides{
				CurrentContext: contextName,
			}

			config, err = clientcmd.NewDefaultClientConfig(*kubeconfig, overrides).ClientConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to build kubeconfig for context '%s': %w", contextName, err)
			}
		} else {
			// Use kubeconfig file with current context
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
			if err != nil {
				return nil, fmt.Errorf("failed to build kubeconfig: %w", err)
			}
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

// Start initialises informer factories, starts all informers, and waits for cache sync.
// Must be called before GetClusterState.
func (k *Client) Start(ctx context.Context) error {
	k.mu.Lock()
	if k.started {
		k.mu.Unlock()
		return fmt.Errorf("kubernetes client already started")
	}
	k.started = true
	k.mu.Unlock()

	// Resync period 0: rely on watch events only. Periodic full-resync would
	// add redundant API server load; missed events are recoverable via re-list
	// on watch reconnect, which the informer machinery handles automatically.
	k.k8sFactory = informers.NewSharedInformerFactory(k.clientset, 0)
	k.istioFactory = istioinformers.NewSharedInformerFactory(k.istioClient, 0)

	// Call .Informer() on each to register them with the factory before Start.
	// The factory only starts informers that have been registered via InformerFor.
	k.k8sFactory.Core().V1().Services().Informer()
	k.k8sFactory.Core().V1().Pods().Informer()
	k.k8sFactory.Discovery().V1().EndpointSlices().Informer()
	k.k8sFactory.Apps().V1().Deployments().Informer()
	k.k8sFactory.Core().V1().Namespaces().Informer()

	k.istioFactory.Networking().V1beta1().DestinationRules().Informer()
	k.istioFactory.Networking().V1beta1().Gateways().Informer()
	k.istioFactory.Networking().V1beta1().Sidecars().Informer()
	k.istioFactory.Networking().V1beta1().VirtualServices().Informer()
	k.istioFactory.Networking().V1beta1().ServiceEntries().Informer()
	k.istioFactory.Networking().V1alpha3().EnvoyFilters().Informer()
	k.istioFactory.Security().V1beta1().RequestAuthentications().Informer()
	k.istioFactory.Security().V1beta1().PeerAuthentications().Informer()
	k.istioFactory.Security().V1beta1().AuthorizationPolicies().Informer()
	k.istioFactory.Extensions().V1alpha1().WasmPlugins().Informer()

	k.k8sFactory.Start(ctx.Done())
	k.istioFactory.Start(ctx.Done())

	k8sSynced := k.k8sFactory.WaitForCacheSync(ctx.Done())
	for t, ok := range k8sSynced {
		if !ok {
			return fmt.Errorf("k8s informer cache sync failed for %v", t)
		}
	}

	istioSynced := k.istioFactory.WaitForCacheSync(ctx.Done())
	for t, ok := range istioSynced {
		if !ok {
			return fmt.Errorf("istio informer cache sync failed for %v", t)
		}
	}

	// InformerFor is idempotent; these calls return the already-registered informers.
	k.servicesLister = k.k8sFactory.Core().V1().Services().Lister()
	k.podsLister = k.k8sFactory.Core().V1().Pods().Lister()
	k.endpointSlicesLister = k.k8sFactory.Discovery().V1().EndpointSlices().Lister()
	k.deploymentsLister = k.k8sFactory.Apps().V1().Deployments().Lister()
	k.namespacesLister = k.k8sFactory.Core().V1().Namespaces().Lister()

	k.destinationRulesLister = k.istioFactory.Networking().V1beta1().DestinationRules().Lister()
	k.gatewaysLister = k.istioFactory.Networking().V1beta1().Gateways().Lister()
	k.sidecarsLister = k.istioFactory.Networking().V1beta1().Sidecars().Lister()
	k.virtualServicesLister = k.istioFactory.Networking().V1beta1().VirtualServices().Lister()
	k.serviceEntriesLister = k.istioFactory.Networking().V1beta1().ServiceEntries().Lister()
	k.envoyFiltersLister = k.istioFactory.Networking().V1alpha3().EnvoyFilters().Lister()
	k.requestAuthenticationsLister = k.istioFactory.Security().V1beta1().RequestAuthentications().Lister()
	k.peerAuthenticationsLister = k.istioFactory.Security().V1beta1().PeerAuthentications().Lister()
	k.authorizationPoliciesLister = k.istioFactory.Security().V1beta1().AuthorizationPolicies().Lister()
	k.wasmPluginsLister = k.istioFactory.Extensions().V1alpha1().WasmPlugins().Lister()

	return nil
}

// GetClientset returns the underlying Kubernetes clientset
func (k *Client) GetClientset() kubernetes.Interface {
	return k.clientset
}

// GetRestConfig returns the underlying Kubernetes REST config
func (k *Client) GetRestConfig() *rest.Config {
	return k.restConfig
}

// GetClusterName retrieves the cluster name from Istio's CLUSTER_ID environment variable in istiod deployment.
// This uses direct API calls and may be called before or after Start.
func (k *Client) GetClusterName(ctx context.Context) (string, error) {
	rootNamespace, activeDeployment := k.discoverIstioControlPlane(ctx)
	if activeDeployment == nil {
		return "", fmt.Errorf("no active istiod deployment found in namespace %s", rootNamespace)
	}

	for _, container := range activeDeployment.Spec.Template.Spec.Containers {
		if container.Name == "discovery" {
			for _, env := range container.Env {
				if env.Name == "CLUSTER_ID" {
					if env.Value != "" {
						k.logger.Debug("found cluster ID from istiod deployment",
							"cluster_id", env.Value,
							"deployment", activeDeployment.Name,
							"namespace", activeDeployment.Namespace)
						return env.Value, nil
					}
				}
			}
		}
	}

	return "", fmt.Errorf("CLUSTER_ID environment variable not found in istiod deployment %s/%s", activeDeployment.Namespace, activeDeployment.Name)
}

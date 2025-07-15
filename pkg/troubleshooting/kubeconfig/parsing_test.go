package kubeconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"github.com/liamawhite/navigator/pkg/envoy/configdump"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to load and parse config dump from configdump testdata
func loadConfigDump(t *testing.T, filename string) map[string]interface{} {
	configPath := filepath.Join("../../envoy/configdump/testdata", filename)
	configDumpBytes, err := os.ReadFile(configPath) //nolint:gosec
	require.NoError(t, err, "failed to read test config dump: %s", configPath)

	var configJSON map[string]interface{}
	err = json.Unmarshal(configDumpBytes, &configJSON)
	require.NoError(t, err, "failed to parse config dump JSON")

	return configJSON
}

// Helper function to load real config dump (which is stored as direct JSON)
func loadRealConfigDump(t *testing.T, filename string) map[string]interface{} {
	configPath := filepath.Join("../../envoy/configdump/testdata", filename)
	configDumpBytes, err := os.ReadFile(configPath) //nolint:gosec
	require.NoError(t, err, "failed to read test config dump: %s", configPath)

	// The real config dump is stored as direct JSON
	var configJSON map[string]interface{}
	err = json.Unmarshal(configDumpBytes, &configJSON)
	require.NoError(t, err, "failed to parse config dump JSON")

	return configJSON
}

func TestConfigDumpParsing(t *testing.T) {
	tests := []struct {
		name              string
		filename          string
		expectedBootstrap bool
		expectedListeners int
		expectedClusters  int
		expectedRoutes    int
		expectedEndpoints int
		validateFunc      func(t *testing.T, parsed *configdump.ParsedConfig)
	}{
		{
			name:              "bootstrap config",
			filename:          "minimal_bootstrap.json",
			expectedBootstrap: true,
			expectedListeners: 0,
			expectedClusters:  0,
			expectedRoutes:    0,
			expectedEndpoints: 0,
			validateFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				assert.Equal(t, "test-node-123", parsed.Bootstrap.GetNode().GetId())
				assert.Equal(t, "test-cluster", parsed.Bootstrap.GetNode().GetCluster())
			},
		},
		{
			name:              "listeners config",
			filename:          "minimal_listeners.json",
			expectedBootstrap: false,
			expectedListeners: 2,
			expectedClusters:  0,
			expectedRoutes:    0,
			expectedEndpoints: 0,
			validateFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				var staticListener, dynamicListener *listenerv3.Listener
				for _, listener := range parsed.Listeners {
					if listener.GetName() == "test-listener" {
						staticListener = listener
					} else if listener.GetName() == "dynamic-listener" {
						dynamicListener = listener
					}
				}
				assert.NotNil(t, staticListener, "should have static listener")
				assert.Equal(t, "test-listener", staticListener.GetName())
				assert.Equal(t, uint32(8080), staticListener.GetAddress().GetSocketAddress().GetPortValue())
				assert.NotNil(t, dynamicListener, "should have dynamic listener")
				assert.Equal(t, "dynamic-listener", dynamicListener.GetName())
				assert.Equal(t, uint32(9090), dynamicListener.GetAddress().GetSocketAddress().GetPortValue())
			},
		},
		{
			name:              "clusters config",
			filename:          "minimal_clusters.json",
			expectedBootstrap: false,
			expectedListeners: 0,
			expectedClusters:  2,
			expectedRoutes:    0,
			expectedEndpoints: 0,
			validateFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				var staticCluster, dynamicCluster *clusterv3.Cluster
				for _, cluster := range parsed.Clusters {
					if cluster.GetName() == "test-static-cluster" {
						staticCluster = cluster
					} else if cluster.GetName() == "test-dynamic-cluster" {
						dynamicCluster = cluster
					}
				}
				assert.NotNil(t, staticCluster, "should have static cluster")
				assert.Equal(t, "test-static-cluster", staticCluster.GetName())
				assert.Equal(t, clusterv3.Cluster_STATIC, staticCluster.GetType())
				assert.NotNil(t, dynamicCluster, "should have dynamic cluster")
				assert.Equal(t, "test-dynamic-cluster", dynamicCluster.GetName())
				assert.Equal(t, clusterv3.Cluster_EDS, dynamicCluster.GetType())
			},
		},
		{
			name:              "routes config",
			filename:          "minimal_routes.json",
			expectedBootstrap: false,
			expectedListeners: 0,
			expectedClusters:  0,
			expectedRoutes:    2,
			expectedEndpoints: 0,
			validateFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				var staticRoute, dynamicRoute *routev3.RouteConfiguration
				for _, route := range parsed.Routes {
					if route.GetName() == "test-static-route" {
						staticRoute = route
					} else if route.GetName() == "test-dynamic-route" {
						dynamicRoute = route
					}
				}
				assert.NotNil(t, staticRoute, "should have static route")
				assert.Equal(t, "test-static-route", staticRoute.GetName())
				assert.Len(t, staticRoute.GetVirtualHosts(), 1)
				assert.NotNil(t, dynamicRoute, "should have dynamic route")
				assert.Equal(t, "test-dynamic-route", dynamicRoute.GetName())
				assert.Len(t, dynamicRoute.GetVirtualHosts(), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configJSON := loadConfigDump(t, tt.filename)
			parser := configdump.NewParser()

			// Convert map back to JSON string for ParseJSON
			configBytes, err := json.Marshal(configJSON)
			require.NoError(t, err)

			parsed, err := parser.ParseJSON(string(configBytes))
			require.NoError(t, err)

			// Check counts
			if tt.expectedBootstrap {
				assert.NotNil(t, parsed.Bootstrap, "bootstrap should be extracted")
			} else {
				assert.Nil(t, parsed.Bootstrap, "bootstrap should not be extracted")
			}
			assert.Len(t, parsed.Listeners, tt.expectedListeners, "unexpected number of listeners")
			assert.Len(t, parsed.Clusters, tt.expectedClusters, "unexpected number of clusters")
			assert.Len(t, parsed.Routes, tt.expectedRoutes, "unexpected number of routes")
			assert.Len(t, parsed.Endpoints, tt.expectedEndpoints, "unexpected number of endpoints")

			// Run additional validation if provided
			if tt.validateFunc != nil {
				tt.validateFunc(t, parsed)
			}
		})
	}
}

func TestRealEnvoyConfigDump(t *testing.T) {
	// This test uses the actual config dump from a running Envoy instance
	configJSON := loadRealConfigDump(t, "envoy_config_dump.json")
	parser := configdump.NewParser()

	// Test the extraction
	// Convert map back to JSON string for ParseJSON
	configBytes, err := json.Marshal(configJSON)
	require.NoError(t, err)

	parsed, err := parser.ParseJSON(string(configBytes))
	require.NoError(t, err, "extraction should not fail")

	tests := []struct {
		name      string
		checkFunc func(t *testing.T, parsed *configdump.ParsedConfig)
	}{
		{
			name: "bootstrap extraction",
			checkFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				// Bootstrap might not be present in all config dumps
				if parsed.Bootstrap != nil {
					t.Logf("Bootstrap node ID: %s", parsed.Bootstrap.GetNode().GetId())
					assert.NotEmpty(t, parsed.Bootstrap.GetNode().GetId(), "bootstrap should have node ID if present")
				} else {
					t.Log("Bootstrap config not present in this dump")
				}
			},
		},
		{
			name: "listeners extraction",
			checkFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				t.Logf("Found %d listeners", len(parsed.Listeners))
				if len(parsed.Listeners) > 0 {
					listenerNames := make([]string, len(parsed.Listeners))
					for i, listener := range parsed.Listeners {
						listenerNames[i] = listener.GetName()
					}
					t.Logf("Listener names: %v", listenerNames)
				} else {
					t.Log("No listeners found in this config dump")
				}
				// Allow empty listeners as some config dumps might not have them
			},
		},
		{
			name: "clusters extraction",
			checkFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				t.Logf("Found %d clusters", len(parsed.Clusters))
				if len(parsed.Clusters) > 0 {
					clusterNames := make([]string, len(parsed.Clusters))
					for i, cluster := range parsed.Clusters {
						clusterNames[i] = cluster.GetName()
					}
					t.Logf("Cluster names: %v", clusterNames)
					// Don't assert on specific cluster names as they vary by config
				} else {
					t.Log("No clusters found in this config dump")
				}
			},
		},
		{
			name: "routes extraction",
			checkFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				assert.Greater(t, len(parsed.Routes), 0, "should extract at least one route")
				if len(parsed.Routes) > 0 {
					routeNames := make([]string, len(parsed.Routes))
					for i, route := range parsed.Routes {
						routeNames[i] = route.GetName()
					}
					t.Logf("Found %d routes: %v", len(parsed.Routes), routeNames)
				}
			},
		},
		{
			name: "endpoints extraction",
			checkFunc: func(t *testing.T, parsed *configdump.ParsedConfig) {
				// Endpoints might be empty in some configs, so we just log what we find
				t.Logf("Found %d endpoints", len(parsed.Endpoints))
				if len(parsed.Endpoints) > 0 {
					endpointNames := make([]string, len(parsed.Endpoints))
					for i, endpoint := range parsed.Endpoints {
						endpointNames[i] = endpoint.GetClusterName()
					}
					t.Logf("Endpoint cluster names: %v", endpointNames)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.checkFunc(t, parsed)
		})
	}
}

func TestParseEnvoyConfigDumpWithMalformedJSON(t *testing.T) {
	tests := []struct {
		name          string
		configJSON    string
		proxyType     string
		version       string
		expectError   bool
		errorContains string
		validateFunc  func(t *testing.T, proxyConfig *v1alpha1.ProxyConfig)
	}{
		{
			name:        "invalid JSON",
			configJSON:  "invalid json",
			proxyType:   "envoy",
			version:     "1.0",
			expectError: false, // parseEnvoyConfigDump now handles errors gracefully
			validateFunc: func(t *testing.T, proxyConfig *v1alpha1.ProxyConfig) {
				// Should return basic config with empty parsing results
				assert.Equal(t, "envoy", proxyConfig.ProxyType)
				assert.Equal(t, "1.0", proxyConfig.Version)
				assert.Equal(t, "invalid json", proxyConfig.RawConfigDump)
				assert.Nil(t, proxyConfig.Bootstrap)
				assert.Empty(t, proxyConfig.Listeners)
			},
		},
		{
			name:        "missing configs",
			configJSON:  `{"not_configs": []}`,
			proxyType:   "envoy",
			version:     "1.0",
			expectError: false,
			validateFunc: func(t *testing.T, proxyConfig *v1alpha1.ProxyConfig) {
				assert.Equal(t, "envoy", proxyConfig.ProxyType)
				assert.Equal(t, "1.0", proxyConfig.Version)
				assert.Empty(t, proxyConfig.Listeners)
				assert.Empty(t, proxyConfig.Clusters)
				assert.Empty(t, proxyConfig.Routes)
				assert.Nil(t, proxyConfig.Bootstrap)
			},
		},
		{
			name:        "empty configs",
			configJSON:  `{"configs": []}`,
			proxyType:   "envoy",
			version:     "1.0",
			expectError: false,
			validateFunc: func(t *testing.T, proxyConfig *v1alpha1.ProxyConfig) {
				assert.Equal(t, "envoy", proxyConfig.ProxyType)
				assert.Equal(t, "1.0", proxyConfig.Version)
				assert.Empty(t, proxyConfig.Listeners)
				assert.Empty(t, proxyConfig.Clusters)
				assert.Empty(t, proxyConfig.Routes)
				assert.Nil(t, proxyConfig.Bootstrap)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &datastore{}
			proxyConfig, err := d.parseEnvoyConfigDump(tt.configJSON, tt.proxyType, tt.version)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, proxyConfig, "proxyConfig should not be nil")
				if tt.validateFunc != nil && proxyConfig != nil {
					tt.validateFunc(t, proxyConfig)
				}
			}
		})
	}
}

// Benchmark tests to ensure parsing performance is acceptable
func BenchmarkParseEnvoyConfigDump(b *testing.B) {
	// Load real config dump from configdump testdata
	configDumpBytes, err := os.ReadFile("../../envoy/configdump/testdata/envoy_config_dump.json")
	if err != nil {
		b.Skip("test config dump not available")
	}

	var configDumpStr string
	err = json.Unmarshal(configDumpBytes, &configDumpStr)
	if err != nil {
		b.Fatalf("failed to parse config dump string: %v", err)
	}

	d := &datastore{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := d.parseEnvoyConfigDump(configDumpStr, "istio-proxy", "1.34.2")
		if err != nil {
			b.Fatalf("parsing failed: %v", err)
		}
	}
}

func BenchmarkConfigdumpParser(b *testing.B) {
	// Load real config dump from configdump testdata
	configDumpBytes, err := os.ReadFile("../../envoy/configdump/testdata/envoy_config_dump.json")
	if err != nil {
		b.Skip("test config dump not available")
	}

	var configDumpStr string
	err = json.Unmarshal(configDumpBytes, &configDumpStr)
	if err != nil {
		b.Fatalf("failed to parse config dump string: %v", err)
	}

	var configJSON map[string]interface{}
	err = json.Unmarshal([]byte(configDumpStr), &configJSON)
	if err != nil {
		b.Fatalf("failed to parse config dump JSON: %v", err)
	}

	parser := configdump.NewParser()

	// Convert map back to JSON string for ParseJSON
	configBytes, err := json.Marshal(configJSON)
	if err != nil {
		b.Fatalf("failed to marshal config: %v", err)
	}
	configStr := string(configBytes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParseJSON(configStr)
		if err != nil {
			b.Fatalf("extraction failed: %v", err)
		}
	}
}

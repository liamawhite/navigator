package configdump

import (
	"os"
	"path/filepath"
	"testing"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to load config dump from testdata as JSON string
func loadConfigDumpString(t *testing.T, filename string) string {
	configPath := filepath.Join("testdata", filename)
	configDumpBytes, err := os.ReadFile(configPath)
	require.NoError(t, err, "failed to read test config dump: %s", configPath)

	return string(configDumpBytes)
}

// Helper function to load real config dump (which is stored as proper JSON)
func loadRealConfigDumpString(t *testing.T, filename string) string {
	configPath := filepath.Join("testdata", filename)
	configDumpBytes, err := os.ReadFile(configPath)
	require.NoError(t, err, "failed to read test config dump: %s", configPath)

	// The real config dump is now stored as proper JSON, so we return it as-is
	return string(configDumpBytes)
}

func TestParser_ParseBootstrapConfig(t *testing.T) {
	configDump := loadConfigDumpString(t, "minimal_bootstrap.json")
	parser := NewParser()

	parsed, err := parser.ParseJSON(configDump)
	require.NoError(t, err)

	// Should extract bootstrap
	assert.NotNil(t, parsed.Bootstrap, "bootstrap should be extracted")
	assert.Equal(t, "test-node-123", parsed.Bootstrap.GetNode().GetId())
	assert.Equal(t, "test-cluster", parsed.Bootstrap.GetNode().GetCluster())

	// Should not extract other components
	assert.Empty(t, parsed.Listeners)
	assert.Empty(t, parsed.Clusters)
	assert.Empty(t, parsed.Endpoints)
	assert.Empty(t, parsed.Routes)
}

func TestParser_ParseListenersConfig(t *testing.T) {
	configDump := loadConfigDumpString(t, "minimal_listeners.json")
	parser := NewParser()

	parsed, err := parser.ParseJSON(configDump)
	require.NoError(t, err)

	// Should extract listeners
	assert.Len(t, parsed.Listeners, 2, "should extract 2 listeners (1 static, 1 dynamic)")

	// Check static listener
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

	// Should not extract other components
	assert.Nil(t, parsed.Bootstrap)
	assert.Empty(t, parsed.Clusters)
	assert.Empty(t, parsed.Endpoints)
	assert.Empty(t, parsed.Routes)
}

func TestParser_ParseClustersConfig(t *testing.T) {
	configDump := loadConfigDumpString(t, "minimal_clusters.json")
	parser := NewParser()

	parsed, err := parser.ParseJSON(configDump)
	require.NoError(t, err)

	// Should extract clusters
	assert.Len(t, parsed.Clusters, 2, "should extract 2 clusters (1 static, 1 dynamic)")

	// Check clusters
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

	// Should not extract other components
	assert.Nil(t, parsed.Bootstrap)
	assert.Empty(t, parsed.Listeners)
	assert.Empty(t, parsed.Endpoints)
	assert.Empty(t, parsed.Routes)
}

func TestParser_ParseRoutesConfig(t *testing.T) {
	configDump := loadConfigDumpString(t, "minimal_routes.json")
	parser := NewParser()

	parsed, err := parser.ParseJSON(configDump)
	require.NoError(t, err)

	// Should extract routes
	assert.Len(t, parsed.Routes, 2, "should extract 2 routes (1 static, 1 dynamic)")

	// Check routes
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

	// Should not extract other components
	assert.Nil(t, parsed.Bootstrap)
	assert.Empty(t, parsed.Listeners)
	assert.Empty(t, parsed.Clusters)
	assert.Empty(t, parsed.Endpoints)
}

func TestParser_ParseRealEnvoyConfigDump(t *testing.T) {
	// This test uses the actual config dump from a running Envoy instance
	configDump := loadRealConfigDumpString(t, "envoy_config_dump.json")
	parser := NewParser()

	// Test the extraction
	parsed, err := parser.ParseJSON(configDump)
	require.NoError(t, err, "extraction should not fail")

	// Verify we extracted components from real data
	t.Run("bootstrap extraction", func(t *testing.T) {
		if parsed.Bootstrap != nil {
			assert.NotEmpty(t, parsed.Bootstrap.GetNode().GetId(), "bootstrap should have node ID if present")
			t.Logf("Bootstrap node ID: %s", parsed.Bootstrap.GetNode().GetId())
		} else {
			t.Log("Bootstrap config not present in this dump")
		}
	})

	t.Run("listeners extraction", func(t *testing.T) {
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
	})

	t.Run("clusters extraction", func(t *testing.T) {
		t.Logf("Found %d clusters", len(parsed.Clusters))
		if len(parsed.Clusters) > 0 {
			clusterNames := make([]string, len(parsed.Clusters))
			for i, cluster := range parsed.Clusters {
				clusterNames[i] = cluster.GetName()
			}
			t.Logf("Cluster names: %v", clusterNames)
		} else {
			t.Log("No clusters found in this config dump")
		}
	})

	t.Run("endpoints extraction", func(t *testing.T) {
		t.Logf("Found %d endpoints", len(parsed.Endpoints))
		if len(parsed.Endpoints) > 0 {
			endpointNames := make([]string, len(parsed.Endpoints))
			for i, endpoint := range parsed.Endpoints {
				endpointNames[i] = endpoint.GetClusterName()
			}
			t.Logf("Endpoint cluster names: %v", endpointNames)
		} else {
			t.Log("No endpoints found in this config dump")
		}
	})

	t.Run("routes extraction", func(t *testing.T) {
		t.Logf("Found %d routes", len(parsed.Routes))
		if len(parsed.Routes) > 0 {
			routeNames := make([]string, len(parsed.Routes))
			for i, route := range parsed.Routes {
				routeNames[i] = route.GetName()
			}
			t.Logf("Route names: %v", routeNames)
		} else {
			t.Log("No routes found in this config dump")
		}
	})
}

func TestParser_ParseInvalidJSON(t *testing.T) {
	parser := NewParser()

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := parser.ParseJSON("invalid json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal config dump")
	})

	t.Run("empty configs", func(t *testing.T) {
		parsed, err := parser.ParseJSON(`{"configs": []}`)
		assert.NoError(t, err)
		assert.Nil(t, parsed.Bootstrap)
		assert.Empty(t, parsed.Listeners)
		assert.Empty(t, parsed.Clusters)
		assert.Empty(t, parsed.Routes)
		assert.Empty(t, parsed.Endpoints)
	})
}

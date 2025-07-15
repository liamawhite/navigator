// Package configdump provides utilities for parsing Envoy configuration dumps.
// It handles the conversion from raw JSON config dumps to structured protobuf types.
package configdump

import (
	"encoding/json"
	"fmt"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	"istio.io/istio/pkg/util/protomarshal"
)

// Well-known filter names from Envoy
const (
	// HTTPConnectionManager network filter
	HTTPConnectionManager = "envoy.filters.network.http_connection_manager"
	// TCPProxy network filter
	TCPProxy = "envoy.filters.network.tcp_proxy"
)

// BlackHoleCluster to catch traffic from routes with unresolved clusters
const BlackHoleCluster = "BlackHoleCluster"

// ParsedConfig represents the parsed components from an Envoy config dump
type ParsedConfig struct {
	Bootstrap *bootstrapv3.Bootstrap
	Listeners []*listenerv3.Listener
	Clusters  []*clusterv3.Cluster
	Endpoints []*endpointv3.ClusterLoadAssignment
	Routes    []*routev3.RouteConfiguration

	// Raw configurations from the original config dump
	RawListeners map[string]string // listener name -> raw JSON config
	RawClusters  map[string]string // cluster name -> raw JSON config
	RawEndpoints map[string]string // endpoint name -> raw JSON config
	RawRoutes    map[string]string // route name -> raw JSON config
}

// ParsedSummary represents the summary components for UI display
type ParsedSummary struct {
	Bootstrap *v1alpha1.BootstrapSummary
	Listeners []*v1alpha1.ListenerSummary
	Clusters  []*v1alpha1.ClusterSummary
	Endpoints []*v1alpha1.EndpointSummary
	Routes    []*v1alpha1.RouteConfigSummary
}

// configDumpWrapper wraps the Envoy ConfigDump with custom unmarshaling
type configDumpWrapper struct {
	*admin.ConfigDump
}

// resolver provides flexible type resolution like istioctl
type resolver struct {
	*protoregistry.Types
}

var nonStrictResolver = &resolver{protoregistry.GlobalTypes}

func (r *resolver) FindMessageByURL(url string) (protoreflect.MessageType, error) {
	typ, err := r.Types.FindMessageByURL(url)
	if err != nil {
		// Ignore unknown types due to Envoy version changes
		msg := exprpb.Type{TypeKind: &exprpb.Type_Dyn{Dyn: &emptypb.Empty{}}}
		return msg.ProtoReflect().Type(), nil
	}
	return typ, nil
}

// UnmarshalJSON provides custom unmarshaling for config dumps
func (w *configDumpWrapper) UnmarshalJSON(b []byte) error {
	cd := &admin.ConfigDump{}
	err := protomarshal.UnmarshalAllowUnknownWithAnyResolver(nonStrictResolver, b, cd)
	*w = configDumpWrapper{cd}
	return err
}

// Parser handles parsing of Envoy configuration dumps
type Parser struct{}

// NewParser creates a new Envoy config dump parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseJSON parses a raw Envoy config dump JSON string into structured protobuf types
func (p *Parser) ParseJSON(rawConfigDump string) (*ParsedConfig, error) {
	// Use istioctl-style parsing with custom unmarshaler
	wrapper := &configDumpWrapper{}
	if err := wrapper.UnmarshalJSON([]byte(rawConfigDump)); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config dump: %w", err)
	}

	return p.parseFromConfigDump(wrapper.ConfigDump)
}

// parseJSONWithRawConfig parses a raw Envoy config dump JSON string and extracts raw config for each listener
func (p *Parser) parseJSONWithRawConfig(rawConfigDump string) (*ParsedConfig, error) {
	// First do the normal parsing
	parsed, err := p.ParseJSON(rawConfigDump)
	if err != nil {
		return nil, err
	}

	// Now extract raw JSON for each listener from the original config dump
	if err := p.extractRawListenerConfigs(rawConfigDump, parsed); err != nil {
		return nil, fmt.Errorf("failed to extract raw listener configs: %w", err)
	}

	return parsed, nil
}

// extractRawListenerConfigs extracts raw JSON configs for each listener from the original config dump
func (p *Parser) extractRawListenerConfigs(rawConfigDump string, parsed *ParsedConfig) error {
	// Parse the raw JSON to extract listener configurations
	var configDump map[string]interface{}
	if err := json.Unmarshal([]byte(rawConfigDump), &configDump); err != nil {
		return fmt.Errorf("failed to unmarshal raw config dump: %w", err)
	}

	// Find the listeners section
	configs, ok := configDump["configs"].([]interface{})
	if !ok {
		return fmt.Errorf("configs section not found or not an array")
	}

	for _, config := range configs {
		configMap, ok := config.(map[string]interface{})
		if !ok {
			continue
		}

		// Look for listener config
		if typeUrl, ok := configMap["@type"].(string); ok && typeUrl == "type.googleapis.com/envoy.admin.v3.ListenersConfigDump" {
			if err := p.extractListenerRawConfigs(configMap, parsed); err != nil {
				return fmt.Errorf("failed to extract listener raw configs: %w", err)
			}
		}
	}

	return nil
}

// extractListenerRawConfigs extracts raw JSON for each individual listener
func (p *Parser) extractListenerRawConfigs(listenersConfigDump map[string]interface{}, parsed *ParsedConfig) error {
	// Extract dynamic listeners
	if dynamicListeners, ok := listenersConfigDump["dynamic_listeners"].([]interface{}); ok {
		for _, dynListener := range dynamicListeners {
			dynListenerMap, ok := dynListener.(map[string]interface{})
			if !ok {
				continue
			}

			if activeState, ok := dynListenerMap["active_state"].(map[string]interface{}); ok {
				if listener, ok := activeState["listener"].(map[string]interface{}); ok {
					if name, ok := listener["name"].(string); ok {
						// Convert back to JSON string
						if rawJSON, err := json.MarshalIndent(listener, "", "  "); err == nil {
							parsed.RawListeners[name] = string(rawJSON)
						}
					}
				}
			}
		}
	}

	// Extract static listeners
	if staticListeners, ok := listenersConfigDump["static_listeners"].([]interface{}); ok {
		for _, staticListener := range staticListeners {
			staticListenerMap, ok := staticListener.(map[string]interface{})
			if !ok {
				continue
			}

			if listener, ok := staticListenerMap["listener"].(map[string]interface{}); ok {
				if name, ok := listener["name"].(string); ok {
					// Convert back to JSON string
					if rawJSON, err := json.MarshalIndent(listener, "", "  "); err == nil {
						parsed.RawListeners[name] = string(rawJSON)
					}
				}
			}
		}
	}

	return nil
}

// ParseJSONToSummary parses a raw Envoy config dump JSON string into summary proto structures
func (p *Parser) ParseJSONToSummary(rawConfigDump string) (*ParsedSummary, error) {
	// First parse to get the structured protobuf types
	parsed, err := p.parseJSONWithRawConfig(rawConfigDump)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config dump: %w", err)
	}

	// Convert to summary structures
	summary := &ParsedSummary{}

	// Convert bootstrap
	if parsed.Bootstrap != nil {
		summary.Bootstrap = p.summarizeBootstrap(parsed.Bootstrap)
	}

	// Convert listeners
	for _, listener := range parsed.Listeners {
		summary.Listeners = append(summary.Listeners, p.summarizeListener(listener, parsed))
	}

	// Convert clusters
	for _, cluster := range parsed.Clusters {
		summary.Clusters = append(summary.Clusters, p.summarizeCluster(cluster))
	}

	// Convert endpoints
	for _, endpoint := range parsed.Endpoints {
		summary.Endpoints = append(summary.Endpoints, p.summarizeEndpoint(endpoint))
	}

	// Convert routes
	for _, route := range parsed.Routes {
		summary.Routes = append(summary.Routes, p.summarizeRouteConfig(route))
	}

	return summary, nil
}

// parseFromConfigDump parses using istioctl-style protobuf unmarshaling
func (p *Parser) parseFromConfigDump(configDump *admin.ConfigDump) (*ParsedConfig, error) {
	parsed := &ParsedConfig{
		RawListeners: make(map[string]string),
		RawClusters:  make(map[string]string),
		RawEndpoints: make(map[string]string),
		RawRoutes:    make(map[string]string),
	}

	// Parse each section using the same approach as istioctl
	for _, config := range configDump.Configs {
		switch config.TypeUrl {
		case "type.googleapis.com/envoy.admin.v3.BootstrapConfigDump":
			if err := p.parseBootstrapFromAny(config, parsed); err != nil {
				// Log error but continue with other configs
				continue
			}

		case "type.googleapis.com/envoy.admin.v3.ListenersConfigDump":
			if err := p.parseListenersFromAny(config, parsed); err != nil {
				// Log error but continue with other configs
				continue
			}

		case "type.googleapis.com/envoy.admin.v3.ClustersConfigDump":
			if err := p.parseClustersFromAny(config, parsed); err != nil {
				// Log error but continue with other configs
				continue
			}

		case "type.googleapis.com/envoy.admin.v3.EndpointsConfigDump":
			if err := p.parseEndpointsFromAny(config, parsed); err != nil {
				// Log error but continue with other configs
				continue
			}

		case "type.googleapis.com/envoy.admin.v3.RoutesConfigDump":
			if err := p.parseRoutesFromAny(config, parsed); err != nil {
				// Log error but continue with other configs
				continue
			}
		}
	}

	return parsed, nil
}

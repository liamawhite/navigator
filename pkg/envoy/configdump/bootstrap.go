package configdump

import (
	"fmt"

	admin "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

// parseBootstrapFromAny extracts bootstrap configuration from protobuf Any
func (p *Parser) parseBootstrapFromAny(configAny *anypb.Any, parsed *ParsedConfig) error {
	bootstrapDump := &admin.BootstrapConfigDump{}
	if err := configAny.UnmarshalTo(bootstrapDump); err != nil {
		return fmt.Errorf("failed to unmarshal bootstrap config dump: %w", err)
	}

	if bootstrapDump.Bootstrap != nil {
		parsed.Bootstrap = bootstrapDump.Bootstrap
	}
	return nil
}

// summarizeBootstrap converts a Bootstrap config to a BootstrapSummary
func (p *Parser) summarizeBootstrap(bootstrap *bootstrapv3.Bootstrap) *v1alpha1.BootstrapSummary {
	if bootstrap == nil {
		return nil
	}

	summary := &v1alpha1.BootstrapSummary{}

	// Extract node information
	if bootstrap.Node != nil {
		summary.Node = &v1alpha1.NodeSummary{
			Id:      bootstrap.Node.Id,
			Cluster: bootstrap.Node.Cluster,
		}

		// Extract metadata as simple string map
		if bootstrap.Node.Metadata != nil && bootstrap.Node.Metadata.Fields != nil {
			summary.Node.Metadata = make(map[string]string)
			for k, v := range bootstrap.Node.Metadata.Fields {
				if v.GetStringValue() != "" {
					summary.Node.Metadata[k] = v.GetStringValue()
				}
			}
		}

		// Extract locality
		if bootstrap.Node.Locality != nil {
			summary.Node.Locality = &v1alpha1.LocalityInfo{
				Region: bootstrap.Node.Locality.Region,
				Zone:   bootstrap.Node.Locality.Zone,
			}
		}
	}

	// Extract admin information
	if bootstrap.Admin != nil && bootstrap.Admin.Address != nil {
		if sockAddr := bootstrap.Admin.Address.GetSocketAddress(); sockAddr != nil {
			summary.AdminAddress = sockAddr.Address
			summary.AdminPort = sockAddr.GetPortValue()
		}
	}

	// Extract dynamic resources information
	if bootstrap.DynamicResources != nil {
		summary.DynamicResourcesConfig = &v1alpha1.DynamicConfigInfo{}

		if ads := bootstrap.DynamicResources.AdsConfig; ads != nil {
			summary.DynamicResourcesConfig.AdsConfig = &v1alpha1.ConfigSourceInfo{
				ConfigSourceSpecifier: "ADS",
			}
		}

		if lds := bootstrap.DynamicResources.LdsConfig; lds != nil {
			summary.DynamicResourcesConfig.LdsConfig = &v1alpha1.ConfigSourceInfo{
				ConfigSourceSpecifier: "LDS",
			}
		}

		if cds := bootstrap.DynamicResources.CdsConfig; cds != nil {
			summary.DynamicResourcesConfig.CdsConfig = &v1alpha1.ConfigSourceInfo{
				ConfigSourceSpecifier: "CDS",
			}
		}
	}

	// Extract cluster manager information
	if bootstrap.ClusterManager != nil {
		summary.ClusterManager = &v1alpha1.ClusterManagerInfo{
			LocalClusterName:   bootstrap.ClusterManager.LocalClusterName,
			OutlierDetection:   bootstrap.ClusterManager.OutlierDetection != nil,
			UpstreamBindConfig: bootstrap.ClusterManager.UpstreamBindConfig != nil,
			LoadStatsConfig:    bootstrap.ClusterManager.LoadStatsConfig != nil,
		}
	}

	return summary
}

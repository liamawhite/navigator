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

package enrich

import (
	"testing"

	v1alpha1 "github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
	"github.com/liamawhite/navigator/pkg/envoy/configdump"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyConfigSummary(t *testing.T) {
	t.Run("enriches proxy config summary", func(t *testing.T) {
		summary := &configdump.ParsedSummary{
			Bootstrap: &v1alpha1.BootstrapSummary{
				Node: &v1alpha1.NodeSummary{Id: "sidecar~10.244.0.1~pod.namespace~cluster.local"},
			},
			Listeners: []*v1alpha1.ListenerSummary{
				{
					Name:    "virtualInbound",
					Address: "0.0.0.0",
					Port:    15006,
				},
			},
			Clusters: []*v1alpha1.ClusterSummary{
				{Name: "outbound|8080||backend.demo.svc.cluster.local"},
			},
			Routes: []*v1alpha1.RouteConfigSummary{
				{Name: "8080"},
			},
		}

		err := ProxyConfigSummary(summary)
		require.NoError(t, err)

		// Check bootstrap enrichment
		assert.Equal(t, v1alpha1.ProxyMode_SIDECAR, summary.Bootstrap.Node.ProxyMode)

		// Check listener enrichment
		assert.Equal(t, v1alpha1.ListenerType_VIRTUAL_INBOUND, summary.Listeners[0].Type)

		// Check cluster enrichment
		assert.Equal(t, v1alpha1.ClusterDirection_OUTBOUND, summary.Clusters[0].Direction)
		assert.Equal(t, uint32(8080), summary.Clusters[0].Port)

		// Check route enrichment
		assert.Equal(t, v1alpha1.RouteType_PORT_BASED, summary.Routes[0].Type)
	})
}

func TestEndpointSummaries(t *testing.T) {
	t.Run("enriches endpoint summaries", func(t *testing.T) {
		endpoints := []*v1alpha1.EndpointSummary{
			{ClusterName: "outbound|8080||backend.demo.svc.cluster.local"},
			{ClusterName: "prometheus_stats"},
		}

		err := EndpointSummaries(endpoints)
		require.NoError(t, err)

		// Check first endpoint enrichment
		assert.Equal(t, v1alpha1.ClusterDirection_OUTBOUND, endpoints[0].Direction)
		assert.Equal(t, uint32(8080), endpoints[0].Port)
		assert.Equal(t, v1alpha1.ClusterType_CLUSTER_EDS, endpoints[0].ClusterType)

		// Check second endpoint enrichment
		assert.Equal(t, v1alpha1.ClusterType_CLUSTER_STATIC, endpoints[1].ClusterType)
	})
}

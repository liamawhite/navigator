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
	"testing"

	"github.com/liamawhite/navigator/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	istioapi "istio.io/api/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClient_convertDestinationRule(t *testing.T) {
	client := &Client{logger: logging.For("test")}

	dr := &istionetworkingv1beta1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dr",
			Namespace: "default",
		},
		Spec: istioapi.DestinationRule{
			Host: "test-service",
			TrafficPolicy: &istioapi.TrafficPolicy{
				LoadBalancer: &istioapi.LoadBalancerSettings{
					LbPolicy: &istioapi.LoadBalancerSettings_Simple{
						Simple: istioapi.LoadBalancerSettings_ROUND_ROBIN,
					},
				},
			},
		},
	}

	result, err := client.convertDestinationRule(dr)

	require.NoError(t, err)
	assert.Equal(t, "test-dr", result.Name)
	assert.Equal(t, "default", result.Namespace)
	assert.Contains(t, result.RawSpec, "test-service")
	assert.Contains(t, result.RawSpec, "ROUND_ROBIN")
}
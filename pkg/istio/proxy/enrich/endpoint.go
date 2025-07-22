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
	"github.com/liamawhite/navigator/pkg/api/types/v1alpha1"
)

// enrichEndpointClusterName parses Istio cluster name information into endpoint summary
func enrichEndpointClusterName() func(*v1alpha1.EndpointSummary) error {
	return func(endpoint *v1alpha1.EndpointSummary) error {
		if endpoint == nil {
			return nil
		}

		ParseClusterName(endpoint.ClusterName, endpoint)
		return nil
	}
}

// enrichEndpointClusterType infers and sets the cluster type based on cluster name patterns
func enrichEndpointClusterType() func(*v1alpha1.EndpointSummary) error {
	return func(endpoint *v1alpha1.EndpointSummary) error {
		if endpoint == nil {
			return nil
		}

		endpoint.ClusterType = InferClusterType(endpoint.ClusterName)
		return nil
	}
}

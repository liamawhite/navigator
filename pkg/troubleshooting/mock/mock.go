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

package mock

import (
	"context"
	"fmt"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
	types "github.com/liamawhite/navigator/pkg/troubleshooting"
)

// Ensure Datastore implements the ProxyDatastore interface
var _ types.ProxyDatastore = (*Datastore)(nil)

// Datastore is a simple mock implementation of ProxyDatastore for testing
type Datastore struct {
	ProxyConfigs map[string]*v1alpha1.ProxyConfig // serviceID:instanceID -> proxy config
}

func (m *Datastore) GetProxyConfig(ctx context.Context, serviceID, instanceID string) (*v1alpha1.ProxyConfig, error) {
	key := fmt.Sprintf("%s:%s", serviceID, instanceID)
	config, exists := m.ProxyConfigs[key]
	if !exists {
		return nil, fmt.Errorf("no proxy config found for service %s instance %s", serviceID, instanceID)
	}
	return config, nil
}

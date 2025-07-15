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

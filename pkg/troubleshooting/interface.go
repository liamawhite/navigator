package troubleshooting

import (
	"context"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

type ProxyDatastore interface {
	GetProxyConfig(ctx context.Context, serviceID, instanceID string) (*v1alpha1.ProxyConfig, error)
}

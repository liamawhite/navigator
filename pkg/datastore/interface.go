package datastore

import (
	"context"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

type ServiceDatastore interface {
	ListServices(ctx context.Context, namespace string) ([]*v1alpha1.Service, error)
	GetService(ctx context.Context, id string) (*v1alpha1.Service, error)
}

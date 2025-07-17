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

package datastore

import (
	"context"

	"github.com/liamawhite/navigator/pkg/api/backend/v1alpha1"
)

type ServiceDatastore interface {
	ListServices(ctx context.Context, namespace string) ([]*v1alpha1.Service, error)
	GetService(ctx context.Context, id string) (*v1alpha1.Service, error)
	GetServiceInstance(ctx context.Context, serviceID, instanceID string) (*v1alpha1.ServiceInstanceDetail, error)
}

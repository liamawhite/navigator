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
	"context"
	"testing"
	"time"

	"github.com/liamawhite/navigator/pkg/logging"
	istiofake "istio.io/client-go/pkg/clientset/versioned/fake"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/require"
)

// newTestClient creates a Client with informers started and caches synced.
// k8sObjects and istioObjects are pre-populated into the respective fake clientsets
// before Start is called, so they are immediately visible through the listers.
func newTestClient(t testing.TB, k8sObjects []runtime.Object, istioObjects []runtime.Object) *Client {
	t.Helper()

	k8sClientset := fake.NewSimpleClientset(k8sObjects...)
	istioClientset := istiofake.NewSimpleClientset(istioObjects...)

	c := &Client{
		clientset:   k8sClientset,
		istioClient: istioClientset,
		logger:      logging.For("test"),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)

	require.NoError(t, c.Start(ctx))
	return c
}

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

package admin

import (
	"bytes"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// KubectlExecInterface defines the interface for executing commands in Kubernetes pods
type KubectlExecInterface interface {
	ExecInContainer(ctx context.Context, namespace, podName, container string, command []string) (string, error)
	HasContainer(ctx context.Context, namespace, podName, container string) (bool, error)
	PodExists(ctx context.Context, namespace, podName string) (bool, error)
}

// KubectlExecClient implements kubectl exec functionality using Kubernetes client-go
type KubectlExecClient struct {
	clientset  kubernetes.Interface
	restConfig *rest.Config
}

// NewKubectlExecClient creates a new kubectl exec client
func NewKubectlExecClient(clientset kubernetes.Interface, restConfig *rest.Config) *KubectlExecClient {
	return &KubectlExecClient{
		clientset:  clientset,
		restConfig: restConfig,
	}
}

// ExecInContainer executes a command in a specific container within a pod
func (k *KubectlExecClient) ExecInContainer(ctx context.Context, namespace, podName, container string, command []string) (string, error) {
	// Create the exec request
	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	// Set up the exec options
	req.VersionedParams(&corev1.PodExecOptions{
		Container: container,
		Command:   command,
		Stdout:    true,
		Stderr:    true,
	}, scheme.ParameterCodec)

	// Create the executor
	exec, err := remotecommand.NewSPDYExecutor(k.restConfig, "POST", req.URL())
	if err != nil {
		return "", fmt.Errorf("failed to create SPDY executor: %w", err)
	}

	// Prepare buffers for stdout and stderr
	var stdout, stderr bytes.Buffer

	// Execute the command
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	if err != nil {
		stderrStr := stderr.String()
		if stderrStr != "" {
			return "", fmt.Errorf("command execution failed: %w, stderr: %s", err, stderrStr)
		}
		return "", fmt.Errorf("command execution failed: %w", err)
	}

	return stdout.String(), nil
}

// HasContainer checks if a pod has a specific container
func (k *KubectlExecClient) HasContainer(ctx context.Context, namespace, podName, container string) (bool, error) {
	pod, err := k.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get pod: %w", err)
	}

	// Check if the container exists in the pod spec
	for _, c := range pod.Spec.Containers {
		if c.Name == container {
			return true, nil
		}
	}

	// Also check init containers
	for _, c := range pod.Spec.InitContainers {
		if c.Name == container {
			return true, nil
		}
	}

	return false, nil
}

// PodExists checks if a pod exists in the given namespace
func (k *KubectlExecClient) PodExists(ctx context.Context, namespace, podName string) (bool, error) {
	_, err := k.clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		// Check if it's a "not found" error
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check pod existence: %w", err)
	}

	return true, nil
}

// MockKubectlExecClient provides a mock implementation for testing
type MockKubectlExecClient struct {
	ExecInContainerFunc func(ctx context.Context, namespace, podName, container string, command []string) (string, error)
	HasContainerFunc    func(ctx context.Context, namespace, podName, container string) (bool, error)
	PodExistsFunc       func(ctx context.Context, namespace, podName string) (bool, error)
}

// ExecInContainer mock implementation
func (m *MockKubectlExecClient) ExecInContainer(ctx context.Context, namespace, podName, container string, command []string) (string, error) {
	if m.ExecInContainerFunc != nil {
		return m.ExecInContainerFunc(ctx, namespace, podName, container, command)
	}
	return "", fmt.Errorf("mock not implemented")
}

// HasContainer mock implementation
func (m *MockKubectlExecClient) HasContainer(ctx context.Context, namespace, podName, container string) (bool, error) {
	if m.HasContainerFunc != nil {
		return m.HasContainerFunc(ctx, namespace, podName, container)
	}
	return false, fmt.Errorf("mock not implemented")
}

// PodExists mock implementation
func (m *MockKubectlExecClient) PodExists(ctx context.Context, namespace, podName string) (bool, error) {
	if m.PodExistsFunc != nil {
		return m.PodExistsFunc(ctx, namespace, podName)
	}
	return false, fmt.Errorf("mock not implemented")
}

// Ensure MockKubectlExecClient implements KubectlExecInterface
var _ KubectlExecInterface = (*MockKubectlExecClient)(nil)

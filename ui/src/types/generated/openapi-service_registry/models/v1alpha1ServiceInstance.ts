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

/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * ServiceInstance represents a single backend instance serving a service.
 */
export type v1alpha1ServiceInstance = {
    instanceId?: string;
    /**
     * ip is the IP address of the instance.
     */
    ip?: string;
    /**
     * pod is the name of the Kubernetes pod backing this instance.
     */
    pod?: string;
    /**
     * namespace is the Kubernetes namespace containing the pod.
     */
    namespace?: string;
    /**
     * cluster_name is the name of the Kubernetes cluster this instance belongs to.
     */
    clusterName?: string;
    /**
     * is_envoy_present indicates whether this instance has an Envoy proxy sidecar.
     */
    isEnvoyPresent?: boolean;
};


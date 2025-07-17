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
import type { v1alpha1ContainerInfo } from './v1alpha1ContainerInfo';
/**
 * ServiceInstanceDetail represents detailed information about a specific service instance.
 */
export type v1alpha1ServiceInstanceDetail = {
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
    /**
     * service_name is the name of the service this instance belongs to.
     */
    serviceName?: string;
    /**
     * pod_status indicates the current status of the pod (e.g., "Running", "Pending", "Failed").
     */
    podStatus?: string;
    /**
     * created_at is the timestamp when the pod was created.
     */
    createdAt?: string;
    /**
     * labels are the Kubernetes labels applied to the pod.
     */
    labels?: Record<string, string>;
    /**
     * annotations are the Kubernetes annotations applied to the pod.
     */
    annotations?: Record<string, string>;
    /**
     * containers lists all containers in the pod.
     */
    containers?: Array<v1alpha1ContainerInfo>;
    /**
     * node_name is the name of the Kubernetes node where the pod is running.
     */
    nodeName?: string;
};


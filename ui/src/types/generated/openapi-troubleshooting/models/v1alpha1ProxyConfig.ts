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
import type { v1alpha1BootstrapSummary } from './v1alpha1BootstrapSummary';
import type { v1alpha1ClusterSummary } from './v1alpha1ClusterSummary';
import type { v1alpha1EndpointSummary } from './v1alpha1EndpointSummary';
import type { v1alpha1ListenerSummary } from './v1alpha1ListenerSummary';
import type { v1alpha1RouteConfigSummary } from './v1alpha1RouteConfigSummary';
/**
 * ProxyConfig represents the configuration of a proxy sidecar (e.g., Envoy).
 */
export type v1alpha1ProxyConfig = {
    /**
     * proxy_type indicates the type of proxy (e.g., "envoy", "istio-proxy").
     */
    proxyType?: string;
    /**
     * version is the version of the proxy software.
     */
    version?: string;
    /**
     * admin_port is the port number where the proxy admin interface is accessible.
     */
    adminPort?: number;
    /**
     * bootstrap contains the bootstrap configuration summary.
     */
    bootstrap?: v1alpha1BootstrapSummary;
    /**
     * listeners contains the listener configuration summaries.
     */
    listeners?: Array<v1alpha1ListenerSummary>;
    /**
     * clusters contains the cluster configuration summaries.
     */
    clusters?: Array<v1alpha1ClusterSummary>;
    /**
     * endpoints contains the endpoint configuration summaries.
     */
    endpoints?: Array<v1alpha1EndpointSummary>;
    /**
     * routes contains the route configuration summaries.
     */
    routes?: Array<v1alpha1RouteConfigSummary>;
    /**
     * raw_config_dump is the original raw configuration dump for debugging.
     */
    rawConfigDump?: string;
};


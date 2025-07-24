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
     * version is the version of the proxy software.
     */
    version?: string;
    /**
     * raw_config_dump is the original raw configuration dump for debugging.
     */
    rawConfigDump?: string;
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
     * raw_clusters is the original raw clusters output from /clusters?format=json endpoint.
     */
    rawClusters?: string;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1AggregatedServicePairMetrics } from './v1alpha1AggregatedServicePairMetrics';
import type { v1alpha1ServicePairMetrics } from './v1alpha1ServicePairMetrics';
/**
 * GetServiceConnectionsResponse contains inbound and outbound service connections.
 */
export type v1alpha1GetServiceConnectionsResponse = {
    /**
     * aggregated_inbound contains properly aggregated metrics for services calling this service.
     */
    aggregatedInbound?: Array<v1alpha1AggregatedServicePairMetrics>;
    /**
     * aggregated_outbound contains properly aggregated metrics for services this service calls.
     */
    aggregatedOutbound?: Array<v1alpha1AggregatedServicePairMetrics>;
    /**
     * detailed_inbound contains per-cluster breakdown for drill-down analysis.
     */
    detailedInbound?: Array<v1alpha1ServicePairMetrics>;
    /**
     * detailed_outbound contains per-cluster breakdown for drill-down analysis.
     */
    detailedOutbound?: Array<v1alpha1ServicePairMetrics>;
    /**
     * timestamp is when these metrics were collected (RFC3339 format).
     */
    timestamp?: string;
    /**
     * clusters_queried lists the clusters that were queried for these metrics.
     */
    clustersQueried?: Array<string>;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1AggregatedServicePairMetrics } from './v1alpha1AggregatedServicePairMetrics';
/**
 * GetServiceConnectionsResponse contains inbound and outbound service connections.
 */
export type v1alpha1GetServiceConnectionsResponse = {
    /**
     * inbound contains aggregated metrics with detailed breakdown for services calling this service.
     */
    inbound?: Array<v1alpha1AggregatedServicePairMetrics>;
    /**
     * outbound contains aggregated metrics with detailed breakdown for services this service calls.
     */
    outbound?: Array<v1alpha1AggregatedServicePairMetrics>;
    /**
     * timestamp is when these metrics were collected (RFC3339 format).
     */
    timestamp?: string;
    /**
     * clusters_queried lists the clusters that were queried for these metrics.
     */
    clustersQueried?: Array<string>;
};


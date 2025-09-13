/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ServicePairMetrics } from './v1alpha1ServicePairMetrics';
/**
 * GetServiceConnectionsResponse contains inbound and outbound service connections.
 */
export type v1alpha1GetServiceConnectionsResponse = {
    /**
     * inbound contains services that call this service.
     */
    inbound?: Array<v1alpha1ServicePairMetrics>;
    /**
     * outbound contains services that this service calls.
     */
    outbound?: Array<v1alpha1ServicePairMetrics>;
    /**
     * timestamp is when these metrics were collected (RFC3339 format).
     */
    timestamp?: string;
    /**
     * clusters_queried lists the clusters that were queried for these metrics.
     */
    clustersQueried?: Array<string>;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ServicePairMetrics } from './v1alpha1ServicePairMetrics';
/**
 * GetServiceGraphMetricsResponse contains service-to-service graph metrics.
 */
export type v1alpha1GetServiceGraphMetricsResponse = {
    /**
     * pairs contains the service-to-service metrics.
     */
    pairs?: Array<v1alpha1ServicePairMetrics>;
    /**
     * timestamp is when these metrics were collected (RFC3339 format).
     */
    timestamp?: string;
    /**
     * clusters_queried lists the clusters that were queried for these metrics.
     */
    clustersQueried?: Array<string>;
};


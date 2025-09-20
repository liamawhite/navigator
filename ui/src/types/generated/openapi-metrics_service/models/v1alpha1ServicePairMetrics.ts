/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1LatencyDistribution } from './v1alpha1LatencyDistribution';
/**
 * ServicePairMetrics represents metrics between a source and destination service.
 */
export type v1alpha1ServicePairMetrics = {
    /**
     * source_cluster is the cluster name of the source service.
     */
    sourceCluster?: string;
    /**
     * source_namespace is the namespace of the source service.
     */
    sourceNamespace?: string;
    /**
     * source_service is the service name of the source service.
     */
    sourceService?: string;
    /**
     * destination_cluster is the cluster name of the destination service.
     */
    destinationCluster?: string;
    /**
     * destination_namespace is the namespace of the destination service.
     */
    destinationNamespace?: string;
    /**
     * destination_service is the service name of the destination service.
     */
    destinationService?: string;
    /**
     * error_rate is the error rate in requests per second.
     */
    errorRate?: number;
    /**
     * request_rate is the request rate in requests per second.
     */
    requestRate?: number;
    /**
     * latency_p99 is the 99th percentile latency.
     */
    latencyP99?: string;
    /**
     * latency_distribution contains the raw histogram distribution for latency.
     * This enables aggregation and percentile calculation at different levels.
     */
    latencyDistribution?: v1alpha1LatencyDistribution;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ClusterPairInfo } from './v1alpha1ClusterPairInfo';
/**
 * AggregatedServicePairMetrics represents properly aggregated metrics across clusters.
 */
export type v1alpha1AggregatedServicePairMetrics = {
    /**
     * source_namespace is the namespace of the source service.
     */
    sourceNamespace?: string;
    /**
     * source_service is the service name of the source service.
     */
    sourceService?: string;
    /**
     * destination_namespace is the namespace of the destination service.
     */
    destinationNamespace?: string;
    /**
     * destination_service is the service name of the destination service.
     */
    destinationService?: string;
    /**
     * error_rate is the aggregated error rate across all clusters.
     */
    errorRate?: number;
    /**
     * request_rate is the aggregated request rate across all clusters.
     */
    requestRate?: number;
    /**
     * latency_p99 is the properly calculated P99 from aggregated histogram.
     */
    latencyP99?: string;
    /**
     * cluster_pairs contains cluster relationship information.
     */
    clusterPairs?: Array<v1alpha1ClusterPairInfo>;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * ClusterPairInfo describes a cluster-to-cluster relationship for a service pair.
 */
export type v1alpha1ClusterPairInfo = {
    /**
     * source_cluster is the cluster name of the source service.
     */
    sourceCluster?: string;
    /**
     * destination_cluster is the cluster name of the destination service.
     */
    destinationCluster?: string;
    /**
     * request_rate is the request rate for this specific cluster pair.
     */
    requestRate?: number;
};


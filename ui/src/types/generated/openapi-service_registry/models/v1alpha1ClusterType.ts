/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * - UNKNOWN_CLUSTER_TYPE: UNKNOWN_CLUSTER_TYPE indicates an unknown or unspecified cluster type
 * - CLUSTER_EDS: CLUSTER_EDS indicates Endpoint Discovery Service clusters (dynamic service discovery)
 * - CLUSTER_STATIC: CLUSTER_STATIC indicates static clusters with predefined endpoints
 * - CLUSTER_STRICT_DNS: CLUSTER_STRICT_DNS indicates clusters using strict DNS resolution
 * - CLUSTER_LOGICAL_DNS: CLUSTER_LOGICAL_DNS indicates clusters using logical DNS resolution
 * - CLUSTER_ORIGINAL_DST: CLUSTER_ORIGINAL_DST indicates clusters using original destination routing
 */
export enum v1alpha1ClusterType {
    UNKNOWN_CLUSTER_TYPE = 'UNKNOWN_CLUSTER_TYPE',
    CLUSTER_EDS = 'CLUSTER_EDS',
    CLUSTER_STATIC = 'CLUSTER_STATIC',
    CLUSTER_STRICT_DNS = 'CLUSTER_STRICT_DNS',
    CLUSTER_LOGICAL_DNS = 'CLUSTER_LOGICAL_DNS',
    CLUSTER_ORIGINAL_DST = 'CLUSTER_ORIGINAL_DST',
}

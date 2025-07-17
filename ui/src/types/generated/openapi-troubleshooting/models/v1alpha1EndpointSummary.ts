/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ClusterDirection } from './v1alpha1ClusterDirection';
import type { v1alpha1ClusterType } from './v1alpha1ClusterType';
import type { v1alpha1EndpointInfo } from './v1alpha1EndpointInfo';
export type v1alpha1EndpointSummary = {
    clusterName?: string;
    endpoints?: Array<v1alpha1EndpointInfo>;
    clusterType?: v1alpha1ClusterType;
    direction?: v1alpha1ClusterDirection;
    port?: number;
    subset?: string;
    serviceFqdn?: string;
};


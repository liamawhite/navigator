/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ClusterDirection } from './v1alpha1ClusterDirection';
export type v1alpha1ClusterSummary = {
    name?: string;
    type?: string;
    connectTimeout?: string;
    loadBalancingPolicy?: string;
    altStatName?: string;
    direction?: v1alpha1ClusterDirection;
    port?: number;
    subset?: string;
    serviceFqdn?: string;
    rawConfig?: string;
};


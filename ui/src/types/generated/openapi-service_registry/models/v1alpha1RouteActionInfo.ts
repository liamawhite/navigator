/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1WeightedClusterInfo } from './v1alpha1WeightedClusterInfo';
export type v1alpha1RouteActionInfo = {
    actionType?: string;
    cluster?: string;
    weightedClusters?: Array<v1alpha1WeightedClusterInfo>;
    timeout?: string;
};


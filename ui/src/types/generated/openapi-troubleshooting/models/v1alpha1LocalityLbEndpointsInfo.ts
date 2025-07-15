/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1LbEndpointInfo } from './v1alpha1LbEndpointInfo';
import type { v1alpha1LocalityInfo } from './v1alpha1LocalityInfo';
export type v1alpha1LocalityLbEndpointsInfo = {
    locality?: v1alpha1LocalityInfo;
    lbEndpoints?: Array<v1alpha1LbEndpointInfo>;
    loadBalancingWeight?: number;
    priority?: number;
    proximity?: number;
};


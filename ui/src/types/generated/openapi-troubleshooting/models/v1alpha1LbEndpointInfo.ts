/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1EndpointDetailsInfo } from './v1alpha1EndpointDetailsInfo';
export type v1alpha1LbEndpointInfo = {
    hostIdentifier?: string;
    endpoint?: v1alpha1EndpointDetailsInfo;
    healthStatus?: string;
    metadata?: Record<string, string>;
    loadBalancingWeight?: number;
};

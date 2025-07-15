/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
export type v1alpha1EndpointInfo = {
    address?: string;
    port?: number;
    health?: string;
    weight?: number;
    priority?: number;
    hostIdentifier?: string;
    metadata?: Record<string, string>;
    loadBalancingWeight?: number;
};

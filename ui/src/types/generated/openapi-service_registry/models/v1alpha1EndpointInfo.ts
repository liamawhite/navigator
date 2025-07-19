/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1AddressType } from './v1alpha1AddressType';
export type v1alpha1EndpointInfo = {
    address?: string;
    port?: number;
    health?: string;
    weight?: number;
    priority?: number;
    hostIdentifier?: string;
    metadata?: Record<string, string>;
    loadBalancingWeight?: number;
    addressType?: v1alpha1AddressType;
};


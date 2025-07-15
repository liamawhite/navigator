/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1RateLimitActionInfo } from './v1alpha1RateLimitActionInfo';
import type { v1alpha1RateLimitDescriptorInfo } from './v1alpha1RateLimitDescriptorInfo';
export type v1alpha1RateLimitInfo = {
    stage?: number;
    disableKey?: string;
    actions?: Array<v1alpha1RateLimitActionInfo>;
    limit?: v1alpha1RateLimitDescriptorInfo;
};

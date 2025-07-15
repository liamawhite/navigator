/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1CorsInfo } from './v1alpha1CorsInfo';
import type { v1alpha1HeaderValueOption } from './v1alpha1HeaderValueOption';
import type { v1alpha1HedgePolicyInfo } from './v1alpha1HedgePolicyInfo';
import type { v1alpha1RateLimitInfo } from './v1alpha1RateLimitInfo';
import type { v1alpha1RetryPolicyInfo } from './v1alpha1RetryPolicyInfo';
import type { v1alpha1RouteInfo } from './v1alpha1RouteInfo';
import type { v1alpha1VirtualClusterInfo } from './v1alpha1VirtualClusterInfo';
export type v1alpha1VirtualHostInfo = {
    name?: string;
    domains?: Array<string>;
    routes?: Array<v1alpha1RouteInfo>;
    requireTls?: string;
    virtualClusters?: Array<v1alpha1VirtualClusterInfo>;
    rateLimits?: Array<v1alpha1RateLimitInfo>;
    requestHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    requestHeadersToRemove?: Array<string>;
    responseHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    responseHeadersToRemove?: Array<string>;
    cors?: v1alpha1CorsInfo;
    includeRequestAttemptCount?: boolean;
    includeAttemptCountInResponse?: boolean;
    retryPolicy?: v1alpha1RetryPolicyInfo;
    hedgePolicy?: v1alpha1HedgePolicyInfo;
};

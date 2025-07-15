/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1HeaderMatcherInfo } from './v1alpha1HeaderMatcherInfo';
import type { v1alpha1RetryBackOffInfo } from './v1alpha1RetryBackOffInfo';
export type v1alpha1RetryPolicyInfo = {
    retryOn?: string;
    numRetries?: number;
    perTryTimeout?: string;
    retryPriority?: string;
    retryHostPredicate?: Array<string>;
    hostSelectionRetryMaxAttempts?: string;
    retriableStatusCodes?: Array<number>;
    retryBackOff?: v1alpha1RetryBackOffInfo;
    retriableHeaders?: Array<v1alpha1HeaderMatcherInfo>;
    retriableRequestHeaders?: Array<v1alpha1HeaderMatcherInfo>;
};

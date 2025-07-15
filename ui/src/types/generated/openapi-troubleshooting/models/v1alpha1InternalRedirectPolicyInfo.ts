/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1InternalRedirectPredicateInfo } from './v1alpha1InternalRedirectPredicateInfo';
export type v1alpha1InternalRedirectPolicyInfo = {
    maxInternalRedirects?: number;
    redirectResponseCodes?: Array<number>;
    predicates?: Array<v1alpha1InternalRedirectPredicateInfo>;
    allowCrossSchemeRedirect?: boolean;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1FilterChainMatchInfo } from './v1alpha1FilterChainMatchInfo';
import type { v1alpha1FilterSummary } from './v1alpha1FilterSummary';
import type { v1alpha1TLSContextInfo } from './v1alpha1TLSContextInfo';
export type v1alpha1FilterChainSummary = {
    name?: string;
    filters?: Array<v1alpha1FilterSummary>;
    match?: v1alpha1FilterChainMatchInfo;
    tlsContext?: v1alpha1TLSContextInfo;
};


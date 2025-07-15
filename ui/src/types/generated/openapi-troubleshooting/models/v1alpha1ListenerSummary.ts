/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1FilterChainSummary } from './v1alpha1FilterChainSummary';
import type { v1alpha1ListenerFilterSummary } from './v1alpha1ListenerFilterSummary';
import type { v1alpha1ListenerType } from './v1alpha1ListenerType';
export type v1alpha1ListenerSummary = {
    name?: string;
    address?: string;
    port?: number;
    filterChains?: Array<v1alpha1FilterChainSummary>;
    type?: v1alpha1ListenerType;
    useOriginalDst?: boolean;
    listenerFilters?: Array<v1alpha1ListenerFilterSummary>;
    rawConfig?: string;
};


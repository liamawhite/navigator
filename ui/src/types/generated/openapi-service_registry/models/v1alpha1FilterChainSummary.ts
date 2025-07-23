/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1FilterInfo } from './v1alpha1FilterInfo';
export type v1alpha1FilterChainSummary = {
    totalChains?: number;
    httpFilters?: Array<v1alpha1FilterInfo>;
    networkFilters?: Array<v1alpha1FilterInfo>;
    tlsContext?: boolean;
};


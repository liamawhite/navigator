/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1LbSubsetSelectorInfo } from './v1alpha1LbSubsetSelectorInfo';
export type v1alpha1LbSubsetConfigInfo = {
    fallbackPolicy?: string;
    defaultSubset?: Record<string, string>;
    subsetSelectors?: Array<v1alpha1LbSubsetSelectorInfo>;
    localityWeightAware?: boolean;
    scaleLocalityWeight?: boolean;
    panicModeAny?: boolean;
    listAsAny?: boolean;
    metadataFallbackPolicy?: string;
    allowRedundantKeys?: boolean;
};


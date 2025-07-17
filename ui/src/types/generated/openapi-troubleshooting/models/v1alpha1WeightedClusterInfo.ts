/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1HeaderValueOption } from './v1alpha1HeaderValueOption';
export type v1alpha1WeightedClusterInfo = {
    name?: string;
    weight?: number;
    metadataMatch?: Record<string, string>;
    requestHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    requestHeadersToRemove?: Array<string>;
    responseHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    responseHeadersToRemove?: Array<string>;
    hostRewriteSpecifier?: string;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1DynamicMetadataInfo } from './v1alpha1DynamicMetadataInfo';
import type { v1alpha1ExtensionInfo } from './v1alpha1ExtensionInfo';
import type { v1alpha1GenericKeyInfo } from './v1alpha1GenericKeyInfo';
import type { v1alpha1HeaderValueMatchInfo } from './v1alpha1HeaderValueMatchInfo';
import type { v1alpha1MetadataInfo } from './v1alpha1MetadataInfo';
import type { v1alpha1RequestHeadersInfo } from './v1alpha1RequestHeadersInfo';
export type v1alpha1RateLimitActionInfo = {
    actionSpecifier?: string;
    sourceCluster?: boolean;
    destinationCluster?: boolean;
    requestHeaders?: v1alpha1RequestHeadersInfo;
    remoteAddress?: boolean;
    genericKey?: v1alpha1GenericKeyInfo;
    headerValueMatch?: v1alpha1HeaderValueMatchInfo;
    dynamicMetadata?: v1alpha1DynamicMetadataInfo;
    metadata?: v1alpha1MetadataInfo;
    extension?: v1alpha1ExtensionInfo;
};


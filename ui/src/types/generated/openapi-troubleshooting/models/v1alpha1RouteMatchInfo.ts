/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1GrpcRouteMatchInfo } from './v1alpha1GrpcRouteMatchInfo';
import type { v1alpha1HeaderMatcherInfo } from './v1alpha1HeaderMatcherInfo';
import type { v1alpha1MetadataMatcherInfo } from './v1alpha1MetadataMatcherInfo';
import type { v1alpha1QueryParameterMatcherInfo } from './v1alpha1QueryParameterMatcherInfo';
import type { v1alpha1RuntimeFractionInfo } from './v1alpha1RuntimeFractionInfo';
import type { v1alpha1TlsContextMatchInfo } from './v1alpha1TlsContextMatchInfo';
export type v1alpha1RouteMatchInfo = {
    pathSpecifier?: string;
    path?: string;
    caseSensitive?: boolean;
    runtimeFraction?: v1alpha1RuntimeFractionInfo;
    headers?: Array<v1alpha1HeaderMatcherInfo>;
    queryParameters?: Array<v1alpha1QueryParameterMatcherInfo>;
    grpc?: v1alpha1GrpcRouteMatchInfo;
    tlsContext?: v1alpha1TlsContextMatchInfo;
    dynamicMetadata?: Array<v1alpha1MetadataMatcherInfo>;
};


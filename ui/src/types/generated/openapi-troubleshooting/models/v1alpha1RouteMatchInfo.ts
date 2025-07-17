// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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


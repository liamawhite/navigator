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


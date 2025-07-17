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
import type { v1alpha1HeaderValueOption } from './v1alpha1HeaderValueOption';
import type { v1alpha1VirtualHostInfo } from './v1alpha1VirtualHostInfo';
export type v1alpha1RouteConfigInfo = {
    name?: string;
    virtualHosts?: Array<v1alpha1VirtualHostInfo>;
    internalOnlyHeaders?: Array<string>;
    responseHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    responseHeadersToRemove?: Array<string>;
    requestHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    requestHeadersToRemove?: Array<string>;
    validateClusters?: boolean;
};


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
import type { v1alpha1HashPolicyConnectionPropertiesInfo } from './v1alpha1HashPolicyConnectionPropertiesInfo';
import type { v1alpha1HashPolicyCookieInfo } from './v1alpha1HashPolicyCookieInfo';
import type { v1alpha1HashPolicyFilterStateInfo } from './v1alpha1HashPolicyFilterStateInfo';
import type { v1alpha1HashPolicyHeaderInfo } from './v1alpha1HashPolicyHeaderInfo';
import type { v1alpha1HashPolicyQueryParameterInfo } from './v1alpha1HashPolicyQueryParameterInfo';
export type v1alpha1HashPolicyInfo = {
    policySpecifier?: string;
    header?: v1alpha1HashPolicyHeaderInfo;
    cookie?: v1alpha1HashPolicyCookieInfo;
    connectionProperties?: v1alpha1HashPolicyConnectionPropertiesInfo;
    queryParameter?: v1alpha1HashPolicyQueryParameterInfo;
    filterState?: v1alpha1HashPolicyFilterStateInfo;
    terminal?: boolean;
};


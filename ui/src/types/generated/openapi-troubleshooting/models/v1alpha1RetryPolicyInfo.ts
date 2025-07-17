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
import type { v1alpha1HeaderMatcherInfo } from './v1alpha1HeaderMatcherInfo';
import type { v1alpha1RetryBackOffInfo } from './v1alpha1RetryBackOffInfo';
export type v1alpha1RetryPolicyInfo = {
    retryOn?: string;
    numRetries?: number;
    perTryTimeout?: string;
    retryPriority?: string;
    retryHostPredicate?: Array<string>;
    hostSelectionRetryMaxAttempts?: string;
    retriableStatusCodes?: Array<number>;
    retryBackOff?: v1alpha1RetryBackOffInfo;
    retriableHeaders?: Array<v1alpha1HeaderMatcherInfo>;
    retriableRequestHeaders?: Array<v1alpha1HeaderMatcherInfo>;
};


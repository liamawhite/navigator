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
import type { v1alpha1DecoratorInfo } from './v1alpha1DecoratorInfo';
import type { v1alpha1HeaderValueOption } from './v1alpha1HeaderValueOption';
import type { v1alpha1HedgePolicyInfo } from './v1alpha1HedgePolicyInfo';
import type { v1alpha1MaxStreamDurationInfo } from './v1alpha1MaxStreamDurationInfo';
import type { v1alpha1RateLimitInfo } from './v1alpha1RateLimitInfo';
import type { v1alpha1RequestMirrorPolicy } from './v1alpha1RequestMirrorPolicy';
import type { v1alpha1RetryPolicyInfo } from './v1alpha1RetryPolicyInfo';
import type { v1alpha1RouteActionInfo } from './v1alpha1RouteActionInfo';
import type { v1alpha1RouteMatchInfo } from './v1alpha1RouteMatchInfo';
import type { v1alpha1TracingInfo } from './v1alpha1TracingInfo';
export type v1alpha1RouteInfo = {
    name?: string;
    match?: v1alpha1RouteMatchInfo;
    action?: v1alpha1RouteActionInfo;
    decorator?: v1alpha1DecoratorInfo;
    requestHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    requestHeadersToRemove?: Array<string>;
    responseHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    responseHeadersToRemove?: Array<string>;
    tracing?: v1alpha1TracingInfo;
    timeout?: string;
    idleTimeout?: string;
    retryPolicy?: v1alpha1RetryPolicyInfo;
    requestMirrorPolicies?: Array<v1alpha1RequestMirrorPolicy>;
    priority?: string;
    rateLimits?: Array<v1alpha1RateLimitInfo>;
    includeVhRateLimits?: boolean;
    hedgePolicy?: v1alpha1HedgePolicyInfo;
    maxStreamDuration?: v1alpha1MaxStreamDurationInfo;
};


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
import type { v1alpha1CorsInfo } from './v1alpha1CorsInfo';
import type { v1alpha1HashPolicyInfo } from './v1alpha1HashPolicyInfo';
import type { v1alpha1HedgePolicyInfo } from './v1alpha1HedgePolicyInfo';
import type { v1alpha1InternalRedirectPolicyInfo } from './v1alpha1InternalRedirectPolicyInfo';
import type { v1alpha1MaxStreamDurationInfo } from './v1alpha1MaxStreamDurationInfo';
import type { v1alpha1RateLimitInfo } from './v1alpha1RateLimitInfo';
import type { v1alpha1RegexRewriteInfo } from './v1alpha1RegexRewriteInfo';
import type { v1alpha1RequestMirrorPolicy } from './v1alpha1RequestMirrorPolicy';
import type { v1alpha1RetryPolicyInfo } from './v1alpha1RetryPolicyInfo';
import type { v1alpha1UpgradeConfigInfo } from './v1alpha1UpgradeConfigInfo';
import type { v1alpha1WeightedClusterInfo } from './v1alpha1WeightedClusterInfo';
export type v1alpha1RouteActionInfo = {
    actionType?: string;
    cluster?: string;
    clusterHeader?: string;
    weightedClusters?: Array<v1alpha1WeightedClusterInfo>;
    clusterNotFoundResponseCode?: string;
    prefixRewrite?: string;
    regexRewrite?: v1alpha1RegexRewriteInfo;
    hostRewriteSpecifier?: string;
    hostRewriteLiteral?: string;
    autoHostRewrite?: boolean;
    autoHostRewriteHeader?: string;
    timeout?: string;
    idleTimeout?: string;
    retryPolicy?: v1alpha1RetryPolicyInfo;
    requestMirrorPolicies?: Array<v1alpha1RequestMirrorPolicy>;
    priority?: string;
    rateLimits?: Array<v1alpha1RateLimitInfo>;
    includeVhRateLimits?: boolean;
    hashPolicy?: Array<v1alpha1HashPolicyInfo>;
    cors?: v1alpha1CorsInfo;
    maxGrpcTimeout?: string;
    grpcTimeoutOffset?: string;
    upgradeConfigs?: Array<v1alpha1UpgradeConfigInfo>;
    internalRedirectPolicy?: v1alpha1InternalRedirectPolicyInfo;
    maxInternalRedirects?: number;
    hedgePolicy?: v1alpha1HedgePolicyInfo;
    maxStreamDuration?: v1alpha1MaxStreamDurationInfo;
};


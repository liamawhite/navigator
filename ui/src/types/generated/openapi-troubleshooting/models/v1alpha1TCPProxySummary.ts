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
import type { v1alpha1AccessLogInfo } from './v1alpha1AccessLogInfo';
import type { v1alpha1HashPolicyInfo } from './v1alpha1HashPolicyInfo';
import type { v1alpha1TunnelingConfigInfo } from './v1alpha1TunnelingConfigInfo';
import type { v1alpha1WeightedClusterInfo } from './v1alpha1WeightedClusterInfo';
export type v1alpha1TCPProxySummary = {
    statPrefix?: string;
    cluster?: string;
    weightedClusters?: Array<v1alpha1WeightedClusterInfo>;
    idleTimeout?: string;
    downstreamIdleTimeout?: string;
    upstreamIdleTimeout?: string;
    accessLog?: Array<v1alpha1AccessLogInfo>;
    maxConnectAttempts?: number;
    hashPolicy?: Array<v1alpha1HashPolicyInfo>;
    tunnelingConfig?: v1alpha1TunnelingConfigInfo;
    maxDownstreamConnectionDuration?: string;
};


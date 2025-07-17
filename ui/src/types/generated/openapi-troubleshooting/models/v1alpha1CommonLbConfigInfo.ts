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
import type { v1alpha1ConsistentHashingLbConfigInfo } from './v1alpha1ConsistentHashingLbConfigInfo';
import type { v1alpha1FractionInfo } from './v1alpha1FractionInfo';
import type { v1alpha1LocalityLbConfigInfo } from './v1alpha1LocalityLbConfigInfo';
import type { v1alpha1ZoneAwareLbConfigInfo } from './v1alpha1ZoneAwareLbConfigInfo';
export type v1alpha1CommonLbConfigInfo = {
    healthyPanicThreshold?: v1alpha1FractionInfo;
    zoneAwareLbConfig?: v1alpha1ZoneAwareLbConfigInfo;
    localityLbConfig?: v1alpha1LocalityLbConfigInfo;
    updateMergeWindow?: string;
    ignoreNewHostsUntilFirstHc?: boolean;
    closeConnectionsOnHostSetChange?: boolean;
    consistentHashingLbConfig?: v1alpha1ConsistentHashingLbConfigInfo;
};


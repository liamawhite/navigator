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
import type { v1alpha1FilterChainSummary } from './v1alpha1FilterChainSummary';
import type { v1alpha1ListenerFilterSummary } from './v1alpha1ListenerFilterSummary';
import type { v1alpha1ListenerType } from './v1alpha1ListenerType';
export type v1alpha1ListenerSummary = {
    name?: string;
    address?: string;
    port?: number;
    filterChains?: Array<v1alpha1FilterChainSummary>;
    type?: v1alpha1ListenerType;
    useOriginalDst?: boolean;
    listenerFilters?: Array<v1alpha1ListenerFilterSummary>;
    rawConfig?: string;
};


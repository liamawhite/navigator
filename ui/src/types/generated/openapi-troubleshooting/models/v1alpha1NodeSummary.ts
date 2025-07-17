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
import type { v1alpha1LocalityInfo } from './v1alpha1LocalityInfo';
import type { v1alpha1ProxyMode } from './v1alpha1ProxyMode';
export type v1alpha1NodeSummary = {
    id?: string;
    cluster?: string;
    metadata?: Record<string, string>;
    locality?: v1alpha1LocalityInfo;
    proxyMode?: v1alpha1ProxyMode;
};


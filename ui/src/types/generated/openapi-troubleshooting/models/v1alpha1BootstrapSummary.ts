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
import type { v1alpha1ClusterManagerInfo } from './v1alpha1ClusterManagerInfo';
import type { v1alpha1DynamicConfigInfo } from './v1alpha1DynamicConfigInfo';
import type { v1alpha1NodeSummary } from './v1alpha1NodeSummary';
export type v1alpha1BootstrapSummary = {
    node?: v1alpha1NodeSummary;
    staticResourcesVersion?: string;
    dynamicResourcesConfig?: v1alpha1DynamicConfigInfo;
    adminPort?: number;
    adminAddress?: string;
    clusterManager?: v1alpha1ClusterManagerInfo;
};


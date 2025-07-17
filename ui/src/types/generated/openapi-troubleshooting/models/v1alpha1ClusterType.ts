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
/**
 * - UNKNOWN_CLUSTER_TYPE: UNKNOWN_CLUSTER_TYPE indicates an unknown or unspecified cluster type
 * - CLUSTER_EDS: CLUSTER_EDS indicates Endpoint Discovery Service clusters (dynamic service discovery)
 * - CLUSTER_STATIC: CLUSTER_STATIC indicates static clusters with predefined endpoints
 * - CLUSTER_STRICT_DNS: CLUSTER_STRICT_DNS indicates clusters using strict DNS resolution
 * - CLUSTER_LOGICAL_DNS: CLUSTER_LOGICAL_DNS indicates clusters using logical DNS resolution
 * - CLUSTER_ORIGINAL_DST: CLUSTER_ORIGINAL_DST indicates clusters using original destination routing
 */
export enum v1alpha1ClusterType {
    UNKNOWN_CLUSTER_TYPE = 'UNKNOWN_CLUSTER_TYPE',
    CLUSTER_EDS = 'CLUSTER_EDS',
    CLUSTER_STATIC = 'CLUSTER_STATIC',
    CLUSTER_STRICT_DNS = 'CLUSTER_STRICT_DNS',
    CLUSTER_LOGICAL_DNS = 'CLUSTER_LOGICAL_DNS',
    CLUSTER_ORIGINAL_DST = 'CLUSTER_ORIGINAL_DST',
}

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
 * - PORT_BASED: PORT_BASED routes are routes with just port numbers (e.g., "80", "443", "15010")
 * - SERVICE_SPECIFIC: SERVICE_SPECIFIC routes are routes with service hostnames and ports (e.g., "backend.demo.svc.cluster.local:8080", external domains from ServiceEntries)
 * - STATIC: STATIC routes are Istio/Envoy internal routing patterns (e.g., "InboundPassthroughCluster", "inbound|8080||")
 */
export enum v1alpha1RouteType {
    PORT_BASED = 'PORT_BASED',
    SERVICE_SPECIFIC = 'SERVICE_SPECIFIC',
    STATIC = 'STATIC',
}

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
import type { v1alpha1CustomHealthCheckInfo } from './v1alpha1CustomHealthCheckInfo';
import type { v1alpha1EventServiceConfigInfo } from './v1alpha1EventServiceConfigInfo';
import type { v1alpha1GrpcHealthCheckInfo } from './v1alpha1GrpcHealthCheckInfo';
import type { v1alpha1HttpHealthCheckInfo } from './v1alpha1HttpHealthCheckInfo';
import type { v1alpha1TcpHealthCheckInfo } from './v1alpha1TcpHealthCheckInfo';
import type { v1alpha1TlsOptionsInfo } from './v1alpha1TlsOptionsInfo';
import type { v1alpha1TransportSocketInfo } from './v1alpha1TransportSocketInfo';
export type v1alpha1HealthCheckInfo = {
    timeout?: string;
    interval?: string;
    intervalJitter?: string;
    intervalJitterPercent?: number;
    unhealthyThreshold?: number;
    healthyThreshold?: number;
    altPort?: number;
    reuseConnection?: boolean;
    httpHealthCheck?: v1alpha1HttpHealthCheckInfo;
    tcpHealthCheck?: v1alpha1TcpHealthCheckInfo;
    grpcHealthCheck?: v1alpha1GrpcHealthCheckInfo;
    customHealthCheck?: v1alpha1CustomHealthCheckInfo;
    noTrafficInterval?: string;
    noTrafficHealthyInterval?: string;
    unhealthyInterval?: string;
    unhealthyEdgeInterval?: string;
    healthyEdgeInterval?: string;
    eventLogPath?: string;
    eventService?: v1alpha1EventServiceConfigInfo;
    alwaysLogHealthCheckFailures?: boolean;
    tlsOptions?: v1alpha1TlsOptionsInfo;
    transportSocket?: v1alpha1TransportSocketInfo;
};


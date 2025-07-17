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
export type v1alpha1OutlierDetectionInfo = {
    consecutiveServerError?: number;
    interval?: string;
    baseEjectionTime?: string;
    maxEjectionPercent?: number;
    minHealthPercent?: number;
    splitExternalLocalOriginErrors?: boolean;
    consecutiveLocalOriginFailure?: number;
    consecutiveGatewayFailure?: number;
    consecutive5xxFailure?: number;
    enforcingConsecutiveServerError?: number;
    enforcingSuccessRate?: number;
    successRateMinimumHosts?: number;
    successRateRequestVolume?: number;
    successRateStdevFactor?: number;
    enforcingConsecutiveLocalOriginFailure?: number;
    enforcingConsecutiveGatewayFailure?: number;
    enforcingLocalOriginSuccessRate?: number;
    localOriginSuccessRateMinimumHosts?: number;
    localOriginSuccessRateRequestVolume?: number;
    localOriginSuccessRateStdevFactor?: number;
    enforcing5xxFailure?: number;
    maxEjectionTime?: string;
};


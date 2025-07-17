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
import type { v1alpha1ConfigSourceInfo } from './v1alpha1ConfigSourceInfo';
export type v1alpha1DynamicConfigInfo = {
    adsConfig?: v1alpha1ConfigSourceInfo;
    ldsConfig?: v1alpha1ConfigSourceInfo;
    cdsConfig?: v1alpha1ConfigSourceInfo;
    edsConfig?: v1alpha1ConfigSourceInfo;
    rdsConfig?: v1alpha1ConfigSourceInfo;
    sdsConfig?: v1alpha1ConfigSourceInfo;
    initialFetchTimeout?: string;
};


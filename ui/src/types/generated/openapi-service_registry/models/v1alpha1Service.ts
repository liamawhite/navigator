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
import type { v1alpha1ServiceInstance } from './v1alpha1ServiceInstance';
/**
 * Service represents a Kubernetes service with its backing instances.
 * Services in different clusters that share the same name and namespace are considered the same service.
 */
export type v1alpha1Service = {
    /**
     * id is a unique identifier for the service in format namespace:service-name (e.g., "default:nginx-service").
     */
    id?: string;
    /**
     * name is the service name.
     */
    name?: string;
    /**
     * namespace is the Kubernetes namespace containing the service.
     */
    namespace?: string;
    /**
     * instances are the backend instances (pods) that serve this service.
     */
    instances?: Array<v1alpha1ServiceInstance>;
};


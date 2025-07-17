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
 * ContainerInfo represents information about a container in a pod.
 */
export type v1alpha1ContainerInfo = {
    /**
     * name is the name of the container.
     */
    name?: string;
    /**
     * image is the container image being used.
     */
    image?: string;
    /**
     * ready indicates whether the container is ready to serve requests.
     */
    ready?: boolean;
    /**
     * restart_count is the number of times the container has been restarted.
     */
    restartCount?: number;
    /**
     * status indicates the current status of the container (e.g., "Running", "Waiting", "Terminated").
     */
    status?: string;
};


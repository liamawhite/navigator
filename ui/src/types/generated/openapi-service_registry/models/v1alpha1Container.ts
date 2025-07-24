/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Container represents a container running in a pod.
 */
export type v1alpha1Container = {
    /**
     * name is the name of the container.
     */
    name?: string;
    /**
     * image is the container image.
     */
    image?: string;
    /**
     * status is the current status of the container (e.g., "Running", "Waiting", "Terminated").
     */
    status?: string;
    /**
     * ready indicates whether the container is ready to serve requests.
     */
    ready?: boolean;
    /**
     * restart_count is the number of times the container has been restarted.
     */
    restartCount?: number;
};


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


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type v1alpha1ListenerDestination = {
    /**
     * destination_type indicates cluster, static IP, original_dst, etc.
     */
    destinationType?: string;
    clusterName?: string;
    address?: string;
    port?: number;
    weight?: number;
    serviceFqdn?: string;
};


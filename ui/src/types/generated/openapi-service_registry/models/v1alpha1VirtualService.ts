/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * VirtualService represents an Istio VirtualService resource.
 */
export type v1alpha1VirtualService = {
    /**
     * name is the name of the virtual service.
     */
    name?: string;
    /**
     * namespace is the namespace of the virtual service.
     */
    namespace?: string;
    /**
     * raw_config is the complete virtual service resource as a JSON string.
     */
    rawConfig?: string;
    /**
     * hosts is the list of destination hosts that these routing rules apply to.
     */
    hosts?: Array<string>;
    /**
     * gateways is the list of gateway names that should apply these routes.
     */
    gateways?: Array<string>;
    /**
     * export_to controls the visibility of this virtual service to other namespaces.
     */
    exportTo?: Array<string>;
};


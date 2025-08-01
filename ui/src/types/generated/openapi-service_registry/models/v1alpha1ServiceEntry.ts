/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * ServiceEntry represents an Istio ServiceEntry resource.
 */
export type v1alpha1ServiceEntry = {
    /**
     * name is the name of the service entry.
     */
    name?: string;
    /**
     * namespace is the namespace of the service entry.
     */
    namespace?: string;
    /**
     * raw_spec is the service entry spec as a JSON string.
     */
    rawSpec?: string;
    /**
     * export_to controls the visibility of this service entry to other namespaces.
     */
    exportTo?: Array<string>;
};


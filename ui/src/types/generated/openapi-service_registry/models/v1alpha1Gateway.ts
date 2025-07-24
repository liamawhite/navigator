/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * Gateway represents an Istio Gateway resource.
 */
export type v1alpha1Gateway = {
    /**
     * name is the name of the gateway.
     */
    name?: string;
    /**
     * namespace is the namespace of the gateway.
     */
    namespace?: string;
    /**
     * raw_spec is the gateway spec as a JSON string.
     */
    rawSpec?: string;
    /**
     * selector is the workload selector for the gateway.
     */
    selector?: Record<string, string>;
};


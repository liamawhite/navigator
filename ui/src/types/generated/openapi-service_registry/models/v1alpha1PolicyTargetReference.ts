/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * PolicyTargetReference represents a reference to a specific resource based on Istio's PolicyTargetReference.
 */
export type v1alpha1PolicyTargetReference = {
    /**
     * group specifies the group of the target resource.
     */
    group?: string;
    /**
     * kind indicates the kind of target resource (required).
     */
    kind?: string;
    /**
     * name provides the name of the target resource (required).
     */
    name?: string;
    /**
     * namespace defines the namespace of the referenced resource.
     * When unspecified, the local namespace is inferred.
     */
    namespace?: string;
};


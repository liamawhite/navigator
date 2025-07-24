/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1PolicyTargetReference } from './v1alpha1PolicyTargetReference';
import type { v1alpha1WorkloadSelector } from './v1alpha1WorkloadSelector';
/**
 * RequestAuthentication represents an Istio RequestAuthentication resource.
 */
export type v1alpha1RequestAuthentication = {
    /**
     * name is the name of the request authentication.
     */
    name?: string;
    /**
     * namespace is the namespace of the request authentication.
     */
    namespace?: string;
    /**
     * raw_spec is the request authentication spec as a JSON string.
     */
    rawSpec?: string;
    /**
     * selector is the criteria used to select the specific set of pods/VMs.
     */
    selector?: v1alpha1WorkloadSelector;
    /**
     * target_refs is the list of resources that this request authentication applies to.
     */
    targetRefs?: Array<v1alpha1PolicyTargetReference>;
};


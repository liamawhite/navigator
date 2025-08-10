/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1PolicyTargetReference } from './v1alpha1PolicyTargetReference';
import type { v1alpha1WorkloadSelector } from './v1alpha1WorkloadSelector';
/**
 * AuthorizationPolicy represents an Istio AuthorizationPolicy resource.
 */
export type v1alpha1AuthorizationPolicy = {
    /**
     * name is the name of the authorization policy.
     */
    name?: string;
    /**
     * namespace is the namespace of the authorization policy.
     */
    namespace?: string;
    /**
     * raw_config is the complete authorization policy resource as a JSON string.
     */
    rawConfig?: string;
    /**
     * selector is the criteria used to select the specific set of pods/VMs.
     */
    selector?: v1alpha1WorkloadSelector;
    /**
     * target_refs is the list of resources that this authorization policy applies to.
     */
    targetRefs?: Array<v1alpha1PolicyTargetReference>;
};


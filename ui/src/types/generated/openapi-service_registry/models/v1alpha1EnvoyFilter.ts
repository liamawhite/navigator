/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1PolicyTargetReference } from './v1alpha1PolicyTargetReference';
import type { v1alpha1WorkloadSelector } from './v1alpha1WorkloadSelector';
/**
 * EnvoyFilter represents an Istio EnvoyFilter resource.
 */
export type v1alpha1EnvoyFilter = {
    /**
     * name is the name of the envoy filter.
     */
    name?: string;
    /**
     * namespace is the namespace of the envoy filter.
     */
    namespace?: string;
    /**
     * raw_spec is the envoy filter spec as a JSON string.
     */
    rawSpec?: string;
    /**
     * workload_selector is the criteria used to select the specific set of pods/VMs.
     */
    workloadSelector?: v1alpha1WorkloadSelector;
    /**
     * target_refs is the list of resources that this envoy filter applies to.
     */
    targetRefs?: Array<v1alpha1PolicyTargetReference>;
};


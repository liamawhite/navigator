/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1WorkloadSelector } from './v1alpha1WorkloadSelector';
/**
 * Sidecar represents an Istio Sidecar resource.
 */
export type v1alpha1Sidecar = {
    /**
     * name is the name of the sidecar.
     */
    name?: string;
    /**
     * namespace is the namespace of the sidecar.
     */
    namespace?: string;
    /**
     * raw_spec is the sidecar spec as a JSON string.
     */
    rawSpec?: string;
    /**
     * workload_selector is the criteria used to select the specific set of pods/VMs.
     */
    workloadSelector?: v1alpha1WorkloadSelector;
};


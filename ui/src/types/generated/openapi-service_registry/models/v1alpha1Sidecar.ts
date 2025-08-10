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
     * raw_config is the complete sidecar resource as a JSON string.
     */
    rawConfig?: string;
    /**
     * workload_selector is the criteria used to select the specific set of pods/VMs.
     */
    workloadSelector?: v1alpha1WorkloadSelector;
};


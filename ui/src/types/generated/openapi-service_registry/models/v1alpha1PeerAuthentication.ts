/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1WorkloadSelector } from './v1alpha1WorkloadSelector';
/**
 * PeerAuthentication represents an Istio PeerAuthentication resource.
 */
export type v1alpha1PeerAuthentication = {
    /**
     * name is the name of the peer authentication.
     */
    name?: string;
    /**
     * namespace is the namespace of the peer authentication.
     */
    namespace?: string;
    /**
     * raw_spec is the peer authentication spec as a JSON string.
     */
    rawSpec?: string;
    /**
     * selector is the criteria used to select the specific set of pods/VMs.
     */
    selector?: v1alpha1WorkloadSelector;
};


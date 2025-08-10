/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1DestinationRuleSubset } from './v1alpha1DestinationRuleSubset';
import type { v1alpha1WorkloadSelector } from './v1alpha1WorkloadSelector';
/**
 * DestinationRule represents an Istio DestinationRule resource.
 */
export type v1alpha1DestinationRule = {
    /**
     * name is the name of the destination rule.
     */
    name?: string;
    /**
     * namespace is the namespace of the destination rule.
     */
    namespace?: string;
    /**
     * raw_config is the complete destination rule resource as a JSON string.
     */
    rawConfig?: string;
    /**
     * host is the name of a service from the service registry.
     */
    host?: string;
    /**
     * subsets is the list of named subsets for traffic routing.
     */
    subsets?: Array<v1alpha1DestinationRuleSubset>;
    /**
     * export_to controls the visibility of this destination rule to other namespaces.
     */
    exportTo?: Array<string>;
    /**
     * workload_selector is the criteria used to select the specific set of pods/VMs.
     */
    workloadSelector?: v1alpha1WorkloadSelector;
};


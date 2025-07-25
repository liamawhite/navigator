/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1PolicyTargetReference } from './v1alpha1PolicyTargetReference';
import type { v1alpha1WorkloadSelector } from './v1alpha1WorkloadSelector';
/**
 * WasmPlugin represents an Istio WasmPlugin resource.
 */
export type v1alpha1WasmPlugin = {
    /**
     * name is the name of the wasm plugin.
     */
    name?: string;
    /**
     * namespace is the namespace of the wasm plugin.
     */
    namespace?: string;
    /**
     * raw_spec is the wasm plugin spec as a JSON string.
     */
    rawSpec?: string;
    /**
     * selector is the criteria used to select the specific set of pods/VMs.
     */
    selector?: v1alpha1WorkloadSelector;
    /**
     * target_refs is the list of resources that this wasm plugin applies to.
     */
    targetRefs?: Array<v1alpha1PolicyTargetReference>;
};


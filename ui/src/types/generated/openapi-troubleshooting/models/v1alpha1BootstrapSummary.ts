/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1ClusterManagerInfo } from './v1alpha1ClusterManagerInfo';
import type { v1alpha1DynamicConfigInfo } from './v1alpha1DynamicConfigInfo';
import type { v1alpha1NodeSummary } from './v1alpha1NodeSummary';
export type v1alpha1BootstrapSummary = {
    node?: v1alpha1NodeSummary;
    staticResourcesVersion?: string;
    dynamicResourcesConfig?: v1alpha1DynamicConfigInfo;
    adminPort?: number;
    adminAddress?: string;
    clusterManager?: v1alpha1ClusterManagerInfo;
};

/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ClusterSpecifierPluginInfo } from './v1alpha1ClusterSpecifierPluginInfo';
import type { v1alpha1HeaderValueOption } from './v1alpha1HeaderValueOption';
import type { v1alpha1VirtualHostInfo } from './v1alpha1VirtualHostInfo';
export type v1alpha1RouteConfigSummary = {
    name?: string;
    virtualHosts?: Array<v1alpha1VirtualHostInfo>;
    internalOnlyHeaders?: Array<string>;
    responseHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    responseHeadersToRemove?: Array<string>;
    requestHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    requestHeadersToRemove?: Array<string>;
    mostSpecificHeaderMutationsWins?: boolean;
    validateClusters?: boolean;
    maxDirectResponseBodySizeBytes?: number;
    clusterSpecifierPlugins?: Array<v1alpha1ClusterSpecifierPluginInfo>;
};


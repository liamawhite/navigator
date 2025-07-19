/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1RouteType } from './v1alpha1RouteType';
import type { v1alpha1VirtualHostInfo } from './v1alpha1VirtualHostInfo';
export type v1alpha1RouteConfigSummary = {
    name?: string;
    virtualHosts?: Array<v1alpha1VirtualHostInfo>;
    internalOnlyHeaders?: Array<string>;
    validateClusters?: boolean;
    rawConfig?: string;
    type?: v1alpha1RouteType;
};


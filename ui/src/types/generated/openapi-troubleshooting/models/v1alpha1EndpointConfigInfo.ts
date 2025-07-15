/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1LocalityLbEndpointsInfo } from './v1alpha1LocalityLbEndpointsInfo';
import type { v1alpha1PolicyInfo } from './v1alpha1PolicyInfo';
export type v1alpha1EndpointConfigInfo = {
    clusterName?: string;
    endpoints?: Array<v1alpha1LocalityLbEndpointsInfo>;
    policy?: v1alpha1PolicyInfo;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1LocalityInfo } from './v1alpha1LocalityInfo';
import type { v1alpha1ProxyMode } from './v1alpha1ProxyMode';
export type v1alpha1NodeSummary = {
    id?: string;
    cluster?: string;
    metadata?: Record<string, string>;
    locality?: v1alpha1LocalityInfo;
    proxyMode?: v1alpha1ProxyMode;
};


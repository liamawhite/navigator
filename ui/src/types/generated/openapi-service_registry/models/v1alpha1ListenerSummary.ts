/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1FilterChainSummary } from './v1alpha1FilterChainSummary';
import type { v1alpha1ListenerRule } from './v1alpha1ListenerRule';
import type { v1alpha1ListenerType } from './v1alpha1ListenerType';
export type v1alpha1ListenerSummary = {
    name?: string;
    address?: string;
    port?: number;
    type?: v1alpha1ListenerType;
    useOriginalDst?: boolean;
    rawConfig?: string;
    rules?: Array<v1alpha1ListenerRule>;
    filterChains?: v1alpha1FilterChainSummary;
};


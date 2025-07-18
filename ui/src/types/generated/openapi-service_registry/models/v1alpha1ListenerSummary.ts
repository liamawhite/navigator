/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ListenerType } from './v1alpha1ListenerType';
export type v1alpha1ListenerSummary = {
    name?: string;
    address?: string;
    port?: number;
    type?: v1alpha1ListenerType;
    useOriginalDst?: boolean;
    rawConfig?: string;
};


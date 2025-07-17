/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1Service } from './v1alpha1Service';
/**
 * ListServicesResponse contains the list of services in the requested namespace(s).
 */
export type v1alpha1ListServicesResponse = {
    /**
     * services is the list of services found in the namespace(s).
     */
    services?: Array<v1alpha1Service>;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1DropOverloadInfo } from './v1alpha1DropOverloadInfo';
export type v1alpha1PolicyInfo = {
    dropOverloads?: Array<v1alpha1DropOverloadInfo>;
    overprovisioningFactor?: number;
    endpointStaleAfter?: string;
    disableOverprovisioning?: boolean;
};

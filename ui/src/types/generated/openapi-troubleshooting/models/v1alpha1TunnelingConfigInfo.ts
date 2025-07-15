/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1HeaderValueOption } from './v1alpha1HeaderValueOption';
export type v1alpha1TunnelingConfigInfo = {
    hostname?: string;
    usePost?: boolean;
    headersToAdd?: Array<v1alpha1HeaderValueOption>;
    propagateResponseHeaders?: Array<string>;
    propagateResponseTrailers?: Array<string>;
};

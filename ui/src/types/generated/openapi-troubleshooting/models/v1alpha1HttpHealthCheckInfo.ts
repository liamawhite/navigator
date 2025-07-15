/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1HeaderValueOption } from './v1alpha1HeaderValueOption';
import type { v1alpha1StatusRangeInfo } from './v1alpha1StatusRangeInfo';
import type { v1alpha1StringMatcherInfo } from './v1alpha1StringMatcherInfo';
export type v1alpha1HttpHealthCheckInfo = {
    host?: string;
    path?: string;
    send?: string;
    receive?: Array<string>;
    requestHeadersToAdd?: Array<v1alpha1HeaderValueOption>;
    requestHeadersToRemove?: Array<string>;
    expectedStatuses?: Array<v1alpha1StatusRangeInfo>;
    useHttp2?: boolean;
    serviceNameMatcher?: v1alpha1StringMatcherInfo;
    method?: string;
};


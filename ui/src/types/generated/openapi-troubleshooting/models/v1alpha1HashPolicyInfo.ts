/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1HashPolicyConnectionPropertiesInfo } from './v1alpha1HashPolicyConnectionPropertiesInfo';
import type { v1alpha1HashPolicyCookieInfo } from './v1alpha1HashPolicyCookieInfo';
import type { v1alpha1HashPolicyFilterStateInfo } from './v1alpha1HashPolicyFilterStateInfo';
import type { v1alpha1HashPolicyHeaderInfo } from './v1alpha1HashPolicyHeaderInfo';
import type { v1alpha1HashPolicyQueryParameterInfo } from './v1alpha1HashPolicyQueryParameterInfo';
export type v1alpha1HashPolicyInfo = {
    policySpecifier?: string;
    header?: v1alpha1HashPolicyHeaderInfo;
    cookie?: v1alpha1HashPolicyCookieInfo;
    connectionProperties?: v1alpha1HashPolicyConnectionPropertiesInfo;
    queryParameter?: v1alpha1HashPolicyQueryParameterInfo;
    filterState?: v1alpha1HashPolicyFilterStateInfo;
    terminal?: boolean;
};


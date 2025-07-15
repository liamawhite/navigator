/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1RangeMatchInfo } from './v1alpha1RangeMatchInfo';
export type v1alpha1HeaderMatcherInfo = {
    name?: string;
    presentMatch?: boolean;
    exactMatch?: string;
    safeRegexMatch?: string;
    rangeMatch?: v1alpha1RangeMatchInfo;
    prefixMatch?: string;
    suffixMatch?: string;
    containsMatch?: string;
    invertMatch?: boolean;
    treatMissingAsEmpty?: boolean;
};


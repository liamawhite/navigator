/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1CustomTagInfo } from './v1alpha1CustomTagInfo';
import type { v1alpha1FractionInfo } from './v1alpha1FractionInfo';
export type v1alpha1TracingInfo = {
    clientSampling?: v1alpha1FractionInfo;
    randomSampling?: v1alpha1FractionInfo;
    overallSampling?: v1alpha1FractionInfo;
    verbose?: boolean;
    maxPathTagLength?: number;
    customTags?: Array<v1alpha1CustomTagInfo>;
};


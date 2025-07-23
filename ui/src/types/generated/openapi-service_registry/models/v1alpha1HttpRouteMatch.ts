/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1HeaderMatchInfo } from './v1alpha1HeaderMatchInfo';
import type { v1alpha1PathMatchInfo } from './v1alpha1PathMatchInfo';
export type v1alpha1HttpRouteMatch = {
    pathMatch?: v1alpha1PathMatchInfo;
    headerMatches?: Array<v1alpha1HeaderMatchInfo>;
    methods?: Array<string>;
};


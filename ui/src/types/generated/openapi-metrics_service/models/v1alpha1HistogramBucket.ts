/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * HistogramBucket represents a single bucket in a histogram distribution.
 */
export type v1alpha1HistogramBucket = {
    /**
     * le is the upper bound of the bucket (less-than-or-equal-to).
     */
    le?: number;
    /**
     * count is the cumulative count of observations in this bucket.
     */
    count?: number;
};


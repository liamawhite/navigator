/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1HistogramBucket } from './v1alpha1HistogramBucket';
/**
 * LatencyDistribution represents a histogram distribution of latency measurements.
 */
export type v1alpha1LatencyDistribution = {
    /**
     * buckets contains the histogram buckets sorted by upper bound.
     */
    buckets?: Array<v1alpha1HistogramBucket>;
    /**
     * total_count is the total number of observations across all buckets.
     */
    totalCount?: number;
    /**
     * sum is the sum of all observed values.
     */
    sum?: number;
};


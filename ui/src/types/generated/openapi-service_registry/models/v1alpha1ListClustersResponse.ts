/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ClusterSyncInfo } from './v1alpha1ClusterSyncInfo';
/**
 * ListClustersResponse contains sync state for all connected clusters.
 */
export type v1alpha1ListClustersResponse = {
    /**
     * clusters is the list of connected clusters with their sync state.
     */
    clusters?: Array<v1alpha1ClusterSyncInfo>;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ClusterSyncInfo } from './v1alpha1ClusterSyncInfo';
/**
 * ListClustersResponse contains the list of all connected clusters and their sync status.
 */
export type v1alpha1ListClustersResponse = {
    /**
     * clusters contains information about each connected cluster's sync status.
     */
    clusters?: Array<v1alpha1ClusterSyncInfo>;
};


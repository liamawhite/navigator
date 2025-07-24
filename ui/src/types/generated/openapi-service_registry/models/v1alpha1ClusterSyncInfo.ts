/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1SyncStatus } from './v1alpha1SyncStatus';
/**
 * ClusterSyncInfo represents the sync state of a connected edge cluster.
 */
export type v1alpha1ClusterSyncInfo = {
    /**
     * cluster_id is the unique identifier for the edge cluster.
     */
    clusterId?: string;
    /**
     * connected_at is when the connection was established (RFC3339 format).
     */
    connectedAt?: string;
    /**
     * last_update is when the last sync occurred (RFC3339 format).
     */
    lastUpdate?: string;
    /**
     * service_count is the number of services currently synced from this cluster.
     */
    serviceCount?: number;
    /**
     * sync_status indicates the health of the sync based on last_update timing.
     */
    syncStatus?: v1alpha1SyncStatus;
};


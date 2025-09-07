/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1SyncStatus } from './v1alpha1SyncStatus';
/**
 * ClusterSyncInfo contains synchronization status and metadata for a connected cluster.
 */
export type v1alpha1ClusterSyncInfo = {
    /**
     * cluster_id uniquely identifies this cluster.
     */
    clusterId?: string;
    /**
     * connected_at is when this cluster initially connected to the manager (RFC3339 format).
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
    /**
     * metrics_enabled indicates whether this cluster's edge supports metrics collection.
     */
    metricsEnabled?: boolean;
};


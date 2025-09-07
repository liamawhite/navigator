/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * SyncStatus represents the health of cluster synchronization.
 *
 * - SYNC_STATUS_INITIALIZING: Connected but hasn't received full state yet
 * - SYNC_STATUS_HEALTHY: Recent updates within expected timeframe
 * - SYNC_STATUS_STALE: No recent updates, potentially problematic
 * - SYNC_STATUS_DISCONNECTED: Connection lost
 */
export enum v1alpha1SyncStatus {
    SYNC_STATUS_UNSPECIFIED = 'SYNC_STATUS_UNSPECIFIED',
    SYNC_STATUS_INITIALIZING = 'SYNC_STATUS_INITIALIZING',
    SYNC_STATUS_HEALTHY = 'SYNC_STATUS_HEALTHY',
    SYNC_STATUS_STALE = 'SYNC_STATUS_STALE',
    SYNC_STATUS_DISCONNECTED = 'SYNC_STATUS_DISCONNECTED',
}

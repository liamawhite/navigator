/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1AccessLogInfo } from './v1alpha1AccessLogInfo';
import type { v1alpha1HashPolicyInfo } from './v1alpha1HashPolicyInfo';
import type { v1alpha1TunnelingConfigInfo } from './v1alpha1TunnelingConfigInfo';
import type { v1alpha1WeightedClusterInfo } from './v1alpha1WeightedClusterInfo';
export type v1alpha1TCPProxySummary = {
    statPrefix?: string;
    cluster?: string;
    weightedClusters?: Array<v1alpha1WeightedClusterInfo>;
    idleTimeout?: string;
    downstreamIdleTimeout?: string;
    upstreamIdleTimeout?: string;
    accessLog?: Array<v1alpha1AccessLogInfo>;
    maxConnectAttempts?: number;
    hashPolicy?: Array<v1alpha1HashPolicyInfo>;
    tunnelingConfig?: v1alpha1TunnelingConfigInfo;
    maxDownstreamConnectionDuration?: string;
};


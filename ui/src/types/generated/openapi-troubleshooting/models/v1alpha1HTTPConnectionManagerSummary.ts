/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1AccessLogInfo } from './v1alpha1AccessLogInfo';
import type { v1alpha1HTTPFilterSummary } from './v1alpha1HTTPFilterSummary';
import type { v1alpha1RDSInfo } from './v1alpha1RDSInfo';
import type { v1alpha1RouteConfigInfo } from './v1alpha1RouteConfigInfo';
export type v1alpha1HTTPConnectionManagerSummary = {
    codecType?: string;
    routeConfig?: v1alpha1RouteConfigInfo;
    rds?: v1alpha1RDSInfo;
    httpFilters?: Array<v1alpha1HTTPFilterSummary>;
    accessLog?: Array<v1alpha1AccessLogInfo>;
    useRemoteAddress?: boolean;
    xffNumTrustedHops?: number;
    skipXffAppend?: boolean;
    via?: string;
    generateRequestId?: boolean;
    forwardClientCertDetails?: string;
    setCurrentClientCertDetails?: boolean;
    proxy100Continue?: boolean;
    streamIdleTimeout?: string;
    requestTimeout?: string;
    drainTimeout?: string;
    delayedCloseTimeout?: string;
    serverName?: string;
};


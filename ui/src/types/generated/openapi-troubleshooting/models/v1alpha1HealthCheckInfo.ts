/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1CustomHealthCheckInfo } from './v1alpha1CustomHealthCheckInfo';
import type { v1alpha1EventServiceConfigInfo } from './v1alpha1EventServiceConfigInfo';
import type { v1alpha1GrpcHealthCheckInfo } from './v1alpha1GrpcHealthCheckInfo';
import type { v1alpha1HttpHealthCheckInfo } from './v1alpha1HttpHealthCheckInfo';
import type { v1alpha1TcpHealthCheckInfo } from './v1alpha1TcpHealthCheckInfo';
import type { v1alpha1TlsOptionsInfo } from './v1alpha1TlsOptionsInfo';
import type { v1alpha1TransportSocketInfo } from './v1alpha1TransportSocketInfo';
export type v1alpha1HealthCheckInfo = {
    timeout?: string;
    interval?: string;
    intervalJitter?: string;
    intervalJitterPercent?: number;
    unhealthyThreshold?: number;
    healthyThreshold?: number;
    altPort?: number;
    reuseConnection?: boolean;
    httpHealthCheck?: v1alpha1HttpHealthCheckInfo;
    tcpHealthCheck?: v1alpha1TcpHealthCheckInfo;
    grpcHealthCheck?: v1alpha1GrpcHealthCheckInfo;
    customHealthCheck?: v1alpha1CustomHealthCheckInfo;
    noTrafficInterval?: string;
    noTrafficHealthyInterval?: string;
    unhealthyInterval?: string;
    unhealthyEdgeInterval?: string;
    healthyEdgeInterval?: string;
    eventLogPath?: string;
    eventService?: v1alpha1EventServiceConfigInfo;
    alwaysLogHealthCheckFailures?: boolean;
    tlsOptions?: v1alpha1TlsOptionsInfo;
    transportSocket?: v1alpha1TransportSocketInfo;
};


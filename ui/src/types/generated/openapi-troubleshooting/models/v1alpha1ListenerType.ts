/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
/**
 * - INBOUND: INBOUND listeners receive traffic from external sources
 * - OUTBOUND: OUTBOUND listeners send traffic to external destinations
 * - VIRTUAL_INBOUND: VIRTUAL_INBOUND listeners are virtual inbound listeners (typically 0.0.0.0 without use_original_dst)
 * - VIRTUAL_OUTBOUND: VIRTUAL_OUTBOUND listeners are virtual outbound listeners (typically 0.0.0.0 with use_original_dst)
 * - METRICS: METRICS listeners serve Prometheus metrics (typically on port 15090)
 * - HEALTHCHECK: HEALTHCHECK listeners serve health check endpoints (typically on port 15021)
 * - ADMIN_XDS: ADMIN_XDS listeners serve Envoy xDS configuration (typically on port 15010)
 * - ADMIN_WEBHOOK: ADMIN_WEBHOOK listeners serve Istio webhook endpoints (typically on port 15012)
 * - ADMIN_DEBUG: ADMIN_DEBUG listeners serve Envoy debug/admin interface (typically on port 15014)
 */
export enum v1alpha1ListenerType {
    INBOUND = 'INBOUND',
    OUTBOUND = 'OUTBOUND',
    VIRTUAL_INBOUND = 'VIRTUAL_INBOUND',
    VIRTUAL_OUTBOUND = 'VIRTUAL_OUTBOUND',
    METRICS = 'METRICS',
    HEALTHCHECK = 'HEALTHCHECK',
    ADMIN_XDS = 'ADMIN_XDS',
    ADMIN_WEBHOOK = 'ADMIN_WEBHOOK',
    ADMIN_DEBUG = 'ADMIN_DEBUG',
}

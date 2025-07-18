/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * - PORT_BASED: PORT_BASED routes are routes with just port numbers (e.g., "80", "443", "15010")
 * - SERVICE_SPECIFIC: SERVICE_SPECIFIC routes are routes with service hostnames and ports (e.g., "backend.demo.svc.cluster.local:8080", external domains from ServiceEntries)
 * - STATIC: STATIC routes are Istio/Envoy internal routing patterns (e.g., "InboundPassthroughCluster", "inbound|8080||")
 */
export enum v1alpha1RouteType {
    PORT_BASED = 'PORT_BASED',
    SERVICE_SPECIFIC = 'SERVICE_SPECIFIC',
    STATIC = 'STATIC',
}

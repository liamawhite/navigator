/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * - UNKNOWN_PROXY_MODE: UNKNOWN_PROXY_MODE indicates an unknown or unspecified proxy mode
 * - SIDECAR: SIDECAR indicates a sidecar proxy (most common in Istio)
 * - GATEWAY: GATEWAY indicates a gateway proxy (ingress/egress gateways)
 * - ROUTER: ROUTER indicates a router proxy
 */
export enum v1alpha1ProxyMode {
    UNKNOWN_PROXY_MODE = 'UNKNOWN_PROXY_MODE',
    SIDECAR = 'SIDECAR',
    GATEWAY = 'GATEWAY',
    ROUTER = 'ROUTER',
}

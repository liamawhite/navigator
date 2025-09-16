/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * - UNKNOWN_PROXY_MODE: UNKNOWN_PROXY_MODE indicates an unknown or unspecified proxy mode
 * - NONE: NONE indicates no proxy is present
 * - SIDECAR: SIDECAR indicates a sidecar proxy (most common in Istio)
 * - ROUTER: ROUTER indicates a router proxy (used for ingress/egress gateways)
 */
export enum v1alpha1ProxyMode {
    UNKNOWN_PROXY_MODE = 'UNKNOWN_PROXY_MODE',
    NONE = 'NONE',
    SIDECAR = 'SIDECAR',
    ROUTER = 'ROUTER',
}

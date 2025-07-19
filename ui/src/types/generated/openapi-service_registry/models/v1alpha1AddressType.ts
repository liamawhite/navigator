/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
/**
 * - UNKNOWN_ADDRESS_TYPE: UNKNOWN_ADDRESS_TYPE indicates an unknown or unspecified address type
 * - SOCKET_ADDRESS: SOCKET_ADDRESS indicates a standard network socket address (IP:port)
 * - ENVOY_INTERNAL_ADDRESS: ENVOY_INTERNAL_ADDRESS indicates an internal Envoy address for listener routing
 * - PIPE_ADDRESS: PIPE_ADDRESS indicates a Unix domain socket address
 */
export enum v1alpha1AddressType {
    UNKNOWN_ADDRESS_TYPE = 'UNKNOWN_ADDRESS_TYPE',
    SOCKET_ADDRESS = 'SOCKET_ADDRESS',
    ENVOY_INTERNAL_ADDRESS = 'ENVOY_INTERNAL_ADDRESS',
    PIPE_ADDRESS = 'PIPE_ADDRESS',
}

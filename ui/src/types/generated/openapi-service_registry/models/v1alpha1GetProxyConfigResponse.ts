/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1ProxyConfig } from './v1alpha1ProxyConfig';
/**
 * GetProxyConfigResponse contains the proxy configuration for the requested pod.
 */
export type v1alpha1GetProxyConfigResponse = {
    /**
     * proxy_config contains the complete Envoy proxy configuration.
     */
    proxyConfig?: v1alpha1ProxyConfig;
};


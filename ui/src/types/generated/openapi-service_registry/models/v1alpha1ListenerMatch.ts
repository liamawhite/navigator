/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1FilterChainMatch } from './v1alpha1FilterChainMatch';
import type { v1alpha1HttpRouteMatch } from './v1alpha1HttpRouteMatch';
import type { v1alpha1TcpProxyMatch } from './v1alpha1TcpProxyMatch';
export type v1alpha1ListenerMatch = {
    httpRoute?: v1alpha1HttpRouteMatch;
    filterChain?: v1alpha1FilterChainMatch;
    tcpProxy?: v1alpha1TcpProxyMatch;
};


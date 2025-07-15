/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1SocketOptionInfo } from './v1alpha1SocketOptionInfo';
import type { v1alpha1TcpKeepaliveInfo } from './v1alpha1TcpKeepaliveInfo';
export type v1alpha1UpstreamConnectionOptionsInfo = {
    tcpKeepalive?: v1alpha1TcpKeepaliveInfo;
    socketOptions?: Array<v1alpha1SocketOptionInfo>;
};


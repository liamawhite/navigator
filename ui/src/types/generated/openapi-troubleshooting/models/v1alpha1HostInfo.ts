/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1HealthCheckConfigInfo } from './v1alpha1HealthCheckConfigInfo';
import type { v1alpha1PipeInfo } from './v1alpha1PipeInfo';
import type { v1alpha1SocketAddressInfo } from './v1alpha1SocketAddressInfo';
export type v1alpha1HostInfo = {
    socketAddress?: v1alpha1SocketAddressInfo;
    pipe?: v1alpha1PipeInfo;
    hostname?: string;
    healthCheckConfig?: v1alpha1HealthCheckConfigInfo;
};

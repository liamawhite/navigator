/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1SocketAddressInfo } from './v1alpha1SocketAddressInfo';
import type { v1alpha1SocketOptionInfo } from './v1alpha1SocketOptionInfo';
export type v1alpha1UpstreamBindConfigInfo = {
    sourceAddress?: v1alpha1SocketAddressInfo;
    freebindInterface?: boolean;
    socketOptions?: Array<v1alpha1SocketOptionInfo>;
};


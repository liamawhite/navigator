/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */

/**
 * ServiceInstance represents a single backend instance serving a service.
 */
export type v1alpha1ServiceInstance = {
    instanceId?: string;
    /**
     * ip is the IP address of the instance.
     */
    ip?: string;
    /**
     * pod is the name of the Kubernetes pod backing this instance.
     */
    pod?: string;
    /**
     * namespace is the Kubernetes namespace containing the pod.
     */
    namespace?: string;
    /**
     * cluster_name is the name of the Kubernetes cluster this instance belongs to.
     */
    clusterName?: string;
    /**
     * is_envoy_present indicates whether this instance has an Envoy proxy sidecar.
     */
    isEnvoyPresent?: boolean;
};

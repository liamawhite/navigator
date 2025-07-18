/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1Container } from './v1alpha1Container';
/**
 * ServiceInstanceDetail represents detailed information about a specific service instance.
 */
export type v1alpha1ServiceInstanceDetail = {
    instanceId?: string;
    /**
     * ip is the IP address of the instance.
     */
    ip?: string;
    /**
     * pod_name is the name of the Kubernetes pod backing this instance.
     */
    podName?: string;
    /**
     * namespace is the Kubernetes namespace containing the pod.
     */
    namespace?: string;
    /**
     * cluster_name is the name of the Kubernetes cluster this instance belongs to.
     */
    clusterName?: string;
    /**
     * envoy_present indicates whether this instance has an Envoy proxy sidecar.
     */
    envoyPresent?: boolean;
    /**
     * service_name is the name of the service this instance belongs to.
     */
    serviceName?: string;
    /**
     * containers is the list of containers running in this pod.
     */
    containers?: Array<v1alpha1Container>;
    /**
     * pod_status is the current status of the pod (e.g., "Running", "Pending").
     */
    podStatus?: string;
    /**
     * node_name is the name of the Kubernetes node hosting this pod.
     */
    nodeName?: string;
    /**
     * created_at is the timestamp when the pod was created.
     */
    createdAt?: string;
    /**
     * labels are the Kubernetes labels assigned to the pod.
     */
    labels?: Record<string, string>;
    /**
     * annotations are the Kubernetes annotations assigned to the pod.
     */
    annotations?: Record<string, string>;
    /**
     * is_envoy_present indicates whether this instance has an Envoy proxy sidecar.
     */
    isEnvoyPresent?: boolean;
};


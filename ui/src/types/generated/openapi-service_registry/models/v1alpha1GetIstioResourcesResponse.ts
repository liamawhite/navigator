/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { v1alpha1DestinationRule } from './v1alpha1DestinationRule';
import type { v1alpha1EnvoyFilter } from './v1alpha1EnvoyFilter';
import type { v1alpha1Gateway } from './v1alpha1Gateway';
import type { v1alpha1PeerAuthentication } from './v1alpha1PeerAuthentication';
import type { v1alpha1Sidecar } from './v1alpha1Sidecar';
import type { v1alpha1VirtualService } from './v1alpha1VirtualService';
/**
 * GetIstioResourcesResponse contains the Istio resources for the requested service instance.
 */
export type v1alpha1GetIstioResourcesResponse = {
    /**
     * virtual_services are VirtualService resources affecting this instance.
     */
    virtualServices?: Array<v1alpha1VirtualService>;
    /**
     * destination_rules are DestinationRule resources affecting this instance.
     */
    destinationRules?: Array<v1alpha1DestinationRule>;
    /**
     * gateways are Gateway resources affecting this instance.
     */
    gateways?: Array<v1alpha1Gateway>;
    /**
     * sidecars are Sidecar resources affecting this instance.
     */
    sidecars?: Array<v1alpha1Sidecar>;
    /**
     * envoy_filters are EnvoyFilter resources affecting this instance.
     */
    envoyFilters?: Array<v1alpha1EnvoyFilter>;
    /**
     * peer_authentications are PeerAuthentication resources affecting this instance.
     */
    peerAuthentications?: Array<v1alpha1PeerAuthentication>;
};


/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
import type { v1alpha1ConsistentHashingLbConfigInfo } from './v1alpha1ConsistentHashingLbConfigInfo';
import type { v1alpha1FractionInfo } from './v1alpha1FractionInfo';
import type { v1alpha1LocalityLbConfigInfo } from './v1alpha1LocalityLbConfigInfo';
import type { v1alpha1ZoneAwareLbConfigInfo } from './v1alpha1ZoneAwareLbConfigInfo';
export type v1alpha1CommonLbConfigInfo = {
    healthyPanicThreshold?: v1alpha1FractionInfo;
    zoneAwareLbConfig?: v1alpha1ZoneAwareLbConfigInfo;
    localityLbConfig?: v1alpha1LocalityLbConfigInfo;
    updateMergeWindow?: string;
    ignoreNewHostsUntilFirstHc?: boolean;
    closeConnectionsOnHostSetChange?: boolean;
    consistentHashingLbConfig?: v1alpha1ConsistentHashingLbConfigInfo;
};

/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
 
export type v1alpha1OutlierDetectionInfo = {
    consecutiveServerError?: number;
    interval?: string;
    baseEjectionTime?: string;
    maxEjectionPercent?: number;
    minHealthPercent?: number;
    splitExternalLocalOriginErrors?: boolean;
    consecutiveLocalOriginFailure?: number;
    consecutiveGatewayFailure?: number;
    consecutive5xxFailure?: number;
    enforcingConsecutiveServerError?: number;
    enforcingSuccessRate?: number;
    successRateMinimumHosts?: number;
    successRateRequestVolume?: number;
    successRateStdevFactor?: number;
    enforcingConsecutiveLocalOriginFailure?: number;
    enforcingConsecutiveGatewayFailure?: number;
    enforcingLocalOriginSuccessRate?: number;
    localOriginSuccessRateMinimumHosts?: number;
    localOriginSuccessRateRequestVolume?: number;
    localOriginSuccessRateStdevFactor?: number;
    enforcing5xxFailure?: number;
    maxEjectionTime?: string;
};


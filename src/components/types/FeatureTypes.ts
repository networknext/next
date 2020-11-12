// Enum to describe different alert types
// Alert types must be a bootstrap 4 valid alert type

export enum FeatureEnum {
  FEATURE_EXPLORE = 'FEATURE_EXPLORE',
  FEATURE_ROUTE_SHADER = 'FEATURE_ROUTE_SHADER',
  FEATURE_IMPERSONATION = 'FEATURE_IMPERSONATION',
  FEATURE_INTERCOM = 'FEATURE_INTERCOM'
}

export interface Flag {
  name: FeatureEnum;
  description: string;
  value: boolean;
}

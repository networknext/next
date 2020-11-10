// Enum to describe different alert types
// Alert types must be a bootstrap 4 valid alert type

export enum FeatureEnums {
  EXPLORE = 'FEATURE_EXPLORE',
  ROUTE_SHADER = 'FEATURE_ROUTE_SHADER',
  IMPERSONATION = 'FEATURE_IMPERSONATION',
  INTERCOM = 'FEATURE_INTERCOM'
}

export interface Flag {
  name: FeatureEnums;
  description: string;
  value: boolean;
}

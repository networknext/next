// Enum to describe different alert types
// Alert types must be a bootstrap 4 valid alert type

export enum FeatureEnum {
  FEATURE_EXPLORE = 'FEATURE_EXPLORE',
  FEATURE_ROUTE_SHADER = 'FEATURE_ROUTE_SHADER',
  FEATURE_IMPERSONATION = 'FEATURE_IMPERSONATION',
  FEATURE_INTERCOM = 'FEATURE_INTERCOM',
  FEATURE_ANALYTICS = 'FEATURE_ANALYTICS',
  FEATURE_TOUR = 'FEATURE_TOUR',
  FEATURE_LOOKER_BIGTABLE_REPLACEMENT = 'FEATURE_LOOKER_BIGTABLE_REPLACEMENT'
}

export interface Flag {
  name: FeatureEnum;
  description: string;
  value: boolean;
}

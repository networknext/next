// Enum to describe different alert types
// Alert types must be a bootstrap 4 valid alert type

export enum DateFilterType {
  LAST_7 = '7 days',
  LAST_14 = '14 days',
  LAST_30 = '30 days',
  LAST_60 = '60 days',
  LAST_90 = '90 days'
}

export interface LookerDateFilterOption {
  name: string;
  value: DateFilterType;
}

export interface Filter {
  companyCode: string;
  dateRange: DateFilterType;
}

// Enum to describe different alert types
// Alert types must be a bootstrap 4 valid alert type

export enum DateFilterType {
  CURRENT_MONTH = 0,
  LAST_MONTH = 1,
  LAST_30 = 2,
  LAST_90 = 3,
  YEAR_TO_DATE = 4,
  CUSTOM = 5
}

export interface Filter {
  companyCode: string;
  dateRange: DateFilterType;
}

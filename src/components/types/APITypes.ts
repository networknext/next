export interface Location {
  continent: string;
  country: string;
  countryCode: string;
  region: string;
  city: string;
  latitude: number;
  longitude: number;
  ISP: string;
  ASN: number;
}

export interface Role {
  ID: string;
  name: string;
  description: string;
}

export interface SessionMeta {
  id: string;
  userHash: string;
  datacenterName: string;
  datacenterAlias: string;
  onNetworkNext: boolean;
  nextRTT: number;
  directRTT: number;
  deltaRTT: number;
  location: Location;
  clientAddr: string;
  serverAddr: string;
  hops: Array<any>; // TODO add a Relay interface for this
  SDK: string;
  connection: string;
  nearbyRelays: Array<any>; // TODO add a Relay interface for this
  platform: string;
  buyerID: string;
}

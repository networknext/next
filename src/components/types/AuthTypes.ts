export interface UserProfile {
  auth0ID: string;
  buyerID: string;
  companyCode: string;
  companyName: string;
  domains: Array<string>;
  email: string;
  idToken: string;
  name: string;
  roles: Array<string>;
  hasAnalytics: boolean;
  hasBilling: boolean;
  hasTrial: boolean;
  verified: boolean;
  routeShader: any;
  pubKey: string;
  newsletterConsent: boolean;
}

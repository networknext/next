export interface UserProfile {
  auth0ID: string;
  buyerID: string;
  seller: boolean;
  companyCode: string;
  companyName: string;
  domains: Array<string>;
  firstName: string;
  lastName: string;
  email: string;
  idToken: string;
  roles: Array<string>;
  verified: boolean;
  routeShader: any;
  pubKey: string;
  newsletterConsent: boolean;
}

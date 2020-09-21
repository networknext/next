export interface UserProfile {
  auth0ID: string;
  companyCode: string;
  companyName: string;
  email: string;
  idToken: string;
  name: string;
  roles: Array<string>;
  verified: boolean;
  routeShader: any;
  domain: string;
  pubKey: string;
  newsletterConsent: boolean
}

export interface UserProfile {
  auth0ID: string;
  avatar: string;
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
  signedTOS: boolean;
  hasAnalytics: boolean;
  hasBilling: boolean;
  hasTrial: boolean;
  verified: boolean;
  routeShader: any;
  pubKey: string;
  newsletterConsent: boolean;
}

export function newDefaultProfile (): UserProfile {
  const defaultProfile: UserProfile = {
    auth0ID: '',
    avatar: '',
    buyerID: '',
    companyCode: '',
    companyName: '',
    domains: [],
    email: '',
    firstName: '',
    hasAnalytics: false,
    hasBilling: false,
    hasTrial: false,
    idToken: '',
    lastName: '',
    newsletterConsent: false,
    pubKey: '',
    roles: [],
    signedTOS: false,
    routeShader: null,
    seller: false,
    verified: false
  }
  return defaultProfile
}

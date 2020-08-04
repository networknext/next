import Auth0Lock from 'auth0-lock'
import store from '@/store'

export default class AuthService {
  // TODO: Make these env vars
  private clientID = 'Kx0mbNIMZtMNA71vf9iatCp3N6qi1GfL'
  private domain = 'networknext.auth0.com'

  public lockClient: Auth0LockStatic

  private getUserInfo: any

  constructor () {
    this.lockClient = new Auth0Lock(
      this.clientID,
      this.domain,
      {
        additionalSignUpFields: [
          {
            name: 'company',
            placeholder: 'Please enter your company name'
          }
        ],
        auth: {
          autoParseHash: true,
          params: {
            scope: 'openid profile email user_metadata app_metadata'
          },
          redirect: false,
          responseType: 'token id_token'
        },
        defaultDatabaseConnection: 'Username-Password-Authentication',
        loginAfterSignUp: true,
        theme: {
          logo: 'https://avatars0.githubusercontent.com/u/31629099?s=200&v=4',
          primaryColor: '#3182bd'
        }
      }
    )
    // HACK - weird build issue is complaining about this on further down so I am doing this for now
    this.getUserInfo = this.lockClient.getUserInfo

    this.lockClient.on('authenticated', this.processAuthentication)
  }

  public signUp () {
    this.lockClient.show({
      allowLogin: false
    })
  }

  public logIn () {
    this.lockClient.show({
      allowSignUp: false
    })
  }

  public logOut () {
    // TODO: Make a env var for baseURL
    this.lockClient.logout({
      returnTo: process.env.VUE_APP_BASE_URL
    })
  }

  private processAuthentication (authResult: AuthResult) {
    this.getUserInfo(authResult.accessToken, (error: auth0.Auth0Error, profile: NNAuth0Profile) => {
      if (!error) {
        const userRoles = profile['https://networknext.com/userRoles'] || { roles: [] }
        const userProfile: UserProfile = {
          auth0ID: profile.sub,
          company: '',
          email: profile.email || '',
          idToken: authResult.idToken,
          name: profile.name,
          roles: userRoles.roles,
          verified: profile.email_verified || false,
          routeShader: null
        }
        store.commit('UPDATE_USER_PROFILE', userProfile)
      }
    })
  }
}

export interface NNAuth0Profile extends auth0.Auth0UserProfile {
  'https://networknext.com/userRoles'?: {
    roles: Array<string>;
  };
}

export interface UserProfile {
  auth0ID: string;
  company: string;
  email: string;
  idToken: string;
  name: string;
  roles: Array<string>;
  verified: boolean;
  routeShader: any;
}

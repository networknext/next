import Auth0Lock from 'auth0-lock'

export default class AuthService {
  // TODO: Make these env vars
  private baseURL = window.location.hostname
  private clientID = this.baseURL === 'portal.networknext.com' ? 'MaSx99ma3AwYOwWMLm3XWNvQ5WyJWG2Y' : 'oQJH3YPHdvZJnxCPo1Irtz5UKi5zrr6n'
  private domain = 'networknext.auth0.com'

  private lockClient: Auth0LockStatic

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
    this.lockClient.on('authenticated', this.processAuthentification)
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
      returnTo: window.location.hostname
    })
  }

  private processAuthentification (authResult: AuthResult) {
    this.getUserInfo(authResult.accessToken, (error: auth0.Auth0Error, profile: any) => {
      if (!error) {
        const roles = profile['https://networknext.com/userRoles'].roles
        const userProfile = {
          auth0ID: profile.sub,
          company: '',
          email: profile.email,
          idToken: authResult.idToken,
          name: profile.name,
          roles: roles
        }
        console.log(userProfile)
      }
    })
  }
}

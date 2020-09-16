import store from '@/store'
import router from '@/router'
import { Auth0Client } from '@auth0/auth0-spa-js'
import { UserProfile } from '@/components/types/AuthTypes.ts'

export class AuthService {
  private clientID: string
  private domain: string
  public authClient: Auth0Client | any

  constructor (options: any) {
    this.clientID = options.clientID
    this.domain = options.domain
    this.authClient = new Auth0Client({
      client_id: this.clientID,
      domain: this.domain,
      cacheLocation: 'localstorage'
    })
    this.processAuthentication()
  }

  public logout () {
    this.authClient.logout()
  }

  public login () {
    this.authClient
      .loginWithPopup({
        redirect_uri: router.currentRoute.fullPath
      })
      .then(() => {
        this.processAuthentication()
      })
      .catch((error: Error) => {
        console.error(error)
      })
  }

  public signUp () {
    this.authClient.loginWithPopup({
      connection: 'Username-Password-Authentication',
      screen_hint: 'signup'
    })
  }

  private async processAuthentication () {
    this.authClient
      .isAuthenticated()
      .then((isAuthenticated: boolean) => {
        if (!isAuthenticated) {
          return
        }
        const userProfile: UserProfile = {
          auth0ID: '',
          companyCode: '',
          companyName: '',
          email: '',
          idToken: '',
          name: '',
          roles: [],
          verified: false,
          routeShader: null,
          domain: '',
          pubKey: '',
          buyerID: '',
          newsletterConsent: false
        }

        this.authClient
          .getIdTokenClaims()
          .then((authResult: any) => {
            const nnScope = authResult[
              'https://networknext.com/userData'
            ]
            const roles: Array<any> = nnScope.roles || { roles: [] }
            const companyCode: string = nnScope.company_code || ''
            const newsletterConsent: boolean = nnScope.newsletterConsent || false
            const email = authResult.email || ''
            const domain = email.split('@')[1]
            const token = authResult.__raw

            userProfile.roles = roles
            userProfile.domain = domain
            userProfile.email = email
            userProfile.idToken = token
            userProfile.auth0ID = authResult.sub
            userProfile.verified = authResult.email_verified
            userProfile.companyCode = companyCode
            userProfile.newsletterConsent = newsletterConsent

            store.commit('UPDATE_USER_PROFILE', userProfile)
          })
          .catch((error: Error) => {
            console.log('Something went wrong fetching user details')
            console.log(error.message)
          })
      })
      .catch((error: Error) => {
        console.log('something went wrong checking auth status')
        console.log(error)
      })
  }
}

export const AuthPlugin = {
  install (Vue: any, options: any) {
    const client = new AuthService({
      domain: options.domain,
      clientID: options.clientID
    })

    Vue.login = () => {
      client.login()
    }

    Vue.logout = () => {
      client.logout()
    }

    Vue.signUp = () => {
      client.signUp()
    }
  }
}

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
      .loginWithRedirect({
        connection: 'Username-Password-Authentication',
        redirect_uri: window.location.origin
      })
  }

  public signUp () {
    this.authClient.loginWithRedirect({
      connection: 'Username-Password-Authentication',
      redirect_uri: window.location.origin + '/?signup=true',
      screen_hint: 'signup'
    })
  }

  public refreshToken () {
    this.authClient.getTokenSilently({ ignoreCache: true })
      .then(() => {
        this.processAuthentication()
      })
  }

  private async processAuthentication () {
    const query = window.location.search

    const isAuthenticated =
      await this.authClient.isAuthenticated()
        .catch((error: Error) => {
          console.log('something went wrong checking auth status')
          console.log(error)
        })
    if (isAuthenticated) {
      const userProfile: UserProfile = {
        auth0ID: '',
        buyerID: '',
        companyCode: '',
        companyName: '',
        email: '',
        idToken: '',
        name: '',
        roles: [],
        verified: false,
        routeShader: null,
        pubKey: '',
        newsletterConsent: false,
        domains: []
      }

      this.authClient
        .getIdTokenClaims()
        .then((authResult: any) => {
          const nnScope = authResult[
            'https://networknext.com/userData'
          ]
          const roles: Array<any> = nnScope.roles || { roles: [] }
          const companyCode: string = nnScope.company_code || ''
          const newsletterConsent: boolean = nnScope.newsletter || false
          const email = authResult.email || ''
          const token = authResult.__raw

          userProfile.roles = roles
          userProfile.email = email
          userProfile.idToken = token
          userProfile.auth0ID = authResult.sub
          userProfile.verified = authResult.email_verified
          userProfile.companyCode = companyCode
          userProfile.newsletterConsent = newsletterConsent

          if (query.includes('signup=true')) {
            store.commit('UPDATE_IS_SIGNUP', true)
          }

          store.commit('UPDATE_USER_PROFILE', userProfile)
        })
        .catch((error: Error) => {
          console.log('Something went wrong fetching user details')
          console.log(error.message)
        })
      return
    }
    // supportSignUp=true&supportForgotPassword=true&email=baumbachandrew%40gmail.com&message=Your email was verified. You can continue using the application.&success=true&code=success
    if (query.includes('code=') && query.includes('state=')) {
      await this.authClient.handleRedirectCallback()
        .catch((error: Error) => {
          console.log('something went wrong with parsing the redirect callback')
          console.log(error)
        })
      this.processAuthentication()
      router.push('/')
    }
  }
}

export const AuthPlugin = {
  install (Vue: any, options: any) {
    Vue.$authService = Vue.prototype.$authService = new AuthService(options)
  }
}

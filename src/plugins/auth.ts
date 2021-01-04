import store from '@/store'
import router from '@/router'
import { Auth0Client } from '@auth0/auth0-spa-js'
import { UserProfile } from '@/components/types/AuthTypes.ts'
import { FeatureEnum } from '@/components/types/FeatureTypes'
import Vue from 'vue'

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
      cacheLocation: 'localstorage',
      useRefreshTokens: true
    })
  }

  public logout () {
    this.authClient.logout({
      returnTo: window.location.origin + '/map'
    })
  }

  public login () {
    this.authClient
      .loginWithRedirect({
        connection: 'Username-Password-Authentication',
        redirect_uri: window.location.origin + '/map'
      })
  }

  public signUp (email: string | undefined) {
    const emailHint = email || ''
    this.authClient.loginWithRedirect({
      connection: 'Username-Password-Authentication',
      redirect_uri: window.location.origin + '/map',
      screen_hint: 'signup',
      login_hint: emailHint
    })
  }

  public refreshToken () {
    return this.authClient.getTokenSilently({ ignoreCache: true })
      .then(() => {
        this.processAuthentication()
      })
  }

  public async processAuthentication (): Promise<any> {
    const query = window.location.search

    if (query.replaceAll('%20', ' ').includes('Your email was verified. You can continue using the application.')) {
      return this.refreshToken()
    }

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

      const authResult = await this.authClient.getIdTokenClaims()
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
      // TODO: There should be a better way to access the Vue instance rather than through the router object
      if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
        (window as any).Intercom('boot', {
          api_base: process.env.VUE_APP_INTERCOM_BASE_API,
          app_id: process.env.VUE_APP_INTERCOM_ID,
          email: email,
          user_id: userProfile.auth0ID,
          unsubscribed_from_emails: newsletterConsent,
          avatar: authResult.picture,
          company: companyCode
        })
      }

      store.commit('UPDATE_USER_PROFILE', userProfile)
      return store.dispatch('processAuthChange')
    }
    if (query.includes('code=') && query.includes('state=')) {
      await this.authClient.handleRedirectCallback()
        .catch((error: Error) => {
          console.log('something went wrong with parsing the redirect callback')
          console.log(error)
        })
      return this.processAuthentication()
    }
    // TODO: There should be a better way to access the Vue instance rather than through the router object
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('boot', {
        api_base: process.env.VUE_APP_INTERCOM_BASE_API,
        app_id: process.env.VUE_APP_INTERCOM_ID
      })
    }
  }
}

export const AuthPlugin = {
  install (Vue: any, options: any) {
    Vue.$authService = Vue.prototype.$authService = new AuthService(options)
  }
}

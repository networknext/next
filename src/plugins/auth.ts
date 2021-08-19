import store from '@/store'
import { Auth0Client } from '@auth0/auth0-spa-js'
import { UserProfile } from '@/components/types/AuthTypes'
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
    // TODO: Redirect should be the page that the user is currently on not defaulting to map
    // IE: User logs in on a session details drill down and redirect to the map?????
    this.authClient
      .loginWithRedirect({
        connection: 'Username-Password-Authentication',
        redirect_uri: window.location.origin + '/map'
      })
  }

  public signUp (email: string | undefined) {
    const emailHint = email || ''
    if (emailHint === '' && Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      console.log('Sending clicked sign up')
      Vue.prototype.$gtag.event('Clicked sign up', {
        event_category: 'Account Creation'
      })
    }
    this.authClient.loginWithRedirect({
      connection: 'Username-Password-Authentication',
      redirect_uri: window.location.origin + '/map?signup=true',
      screen_hint: 'signup',
      login_hint: emailHint
    })
  }

  // TODO: This should be an async function instead of the weird nested promise
  public refreshToken (): Promise<any> {
    return this.authClient.getTokenSilently({ ignoreCache: true })
      .then(() => {
        this.processAuthentication()
      })
  }

  public async processAuthentication (): Promise<any> {
    const query = window.location.search

    // Current version of chrome doesn't like replace all apparently
    if (query.includes('Your%20email%20was%20verified.%20You%20can%20continue%20using%20the%20application.')) {
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
        seller: false,
        companyCode: '',
        companyName: '',
        firstName: '',
        lastName: '',
        email: '',
        idToken: '',
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

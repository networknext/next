import store from '@/store'
import { Auth0DecodedHash, Auth0Error, WebAuth } from 'auth0-js'
import { FeatureEnum } from '@/components/types/FeatureTypes'
import Vue from 'vue'

export class AuthService {
  private clientID: string
  private domain: string
  public auth0Client: WebAuth

  constructor (options: any) {
    this.clientID = options.clientID
    this.domain = options.domain

    this.auth0Client = new WebAuth({
      domain: this.domain,
      clientID: this.clientID,
      responseType: 'id_token',
      redirectUri: window.location.origin + '/map'
    })
  }

  public logout () {
    this.auth0Client.logout({
      returnTo: window.location.origin + '/map'
    })
  }

  public login (username: string, password: string): Promise<any> {
    return new Promise(
      (resolve: any, reject: any) => this.auth0Client.login(
        {
          username: username,
          password: password,
          realm: 'Username-Password-Authentication'
        },
        (err: Auth0Error | null) => {
          err ? reject(Error('Wrong username/password')) : resolve()
        }
      )
    )
  }

  public getAccess (email: string, password: string): Promise<Error | undefined> {
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      Vue.prototype.$gtag.event('clicked sign up', {
        event_category: 'Account Creation',
        event_label: 'Sign up'
      })
    }
    // TODO: this.auth0Client.signupAndAuthorize doesn't work here for some reason
    return new Promise((resolve: any, reject: any) => this.auth0Client.signup({
      username: email,
      email: email,
      password: password,
      connection: 'Username-Password-Authentication'
    }, (err: Auth0Error | null) => {
      err ? reject(new Error('Auth0 failed to sign up user')) : resolve()
    }))
  }

  // TODO: This should be an async function instead of the weird nested promise
  public refreshToken (): Promise<any> {
    return this.processAuthentication()
  }

  public async processAuthentication (): Promise<any> {
    // Auth0 sucks so this is a hack to make the callback of check session resolve as a promise -> undefined if the user is logged out of the result of fetching the local token
    const authResult: Auth0DecodedHash = await new Promise((resolve: any, reject: any) => this.auth0Client.checkSession({}, (err: Auth0Error | null, result: Auth0DecodedHash) => {
      if (err) {
        resolve(undefined)
      } else {
        resolve(result)
      }
    }))

    // If user is logged in process the login
    if (authResult) {
      return store.dispatch('processAuthChange', authResult)
    }

    const isReturning = localStorage.returningUser || 'false'
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_TOUR)) {
      if (!(isReturning === 'true') && store.getters.isAnonymous) {
        store.commit('TOGGLE_IS_TOUR', true)
        localStorage.returningUser = true
      }
    }

    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_INTERCOM)) {
      (window as any).Intercom('boot', {
        api_base: process.env.VUE_APP_INTERCOM_BASE_API,
        app_id: process.env.VUE_APP_INTERCOM_ID
      })
    }

    // Generic resolve to still return a promise even though there isn't anything to resolve here
    return Promise.resolve()
  }
}

export const AuthPlugin = {
  install (Vue: any, options: any) {
    Vue.$authService = Vue.prototype.$authService = new AuthService(options)
  }
}

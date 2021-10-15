import store from '@/store'
import { Auth0Client } from '@auth0/auth0-spa-js'
import { Auth0DecodedHash, Auth0Error, WebAuth } from 'auth0-js'
import { FeatureEnum } from '@/components/types/FeatureTypes'
import Vue from 'vue'

export class AuthService {
  private clientID: string
  private domain: string
  public authClient: Auth0Client
  public customClient: WebAuth

  constructor (options: any) {
    this.clientID = options.clientID
    this.domain = options.domain
    this.authClient = new Auth0Client({
      client_id: this.clientID,
      domain: this.domain,
      cacheLocation: 'localstorage',
      useRefreshTokens: true
    })

    this.customClient = new WebAuth({
      domain: this.domain,
      clientID: this.clientID,
      responseType: 'id_token',
      redirectUri: window.location.origin + '/map'
    })
  }

  public logout () {
    this.customClient.logout({
      returnTo: window.location.origin + '/map'
    })
  }

  public login (username: string, password: string): Promise<any> {
    return new Promise(
      (resolve: any, reject: any) => this.customClient.login(
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

  public getAccess (firstName: string, lastName: string, email: string, password: string, companyName: string, companyWebsite: string) {
    if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
      Vue.prototype.$gtag.event('clicked sign up', {
        event_category: 'Account Creation',
        event_label: 'Sign up'
      })
    }
    // TODO: this.customClient.signupAndAuthorize doesn't work here for some reason
    const signUpPromise = new Promise((resolve: any, reject: any) => this.customClient.signup({
      username: email,
      email: email,
      password: password,
      connection: 'Username-Password-Authentication'
    }, (err: Auth0Error | null) => {
      err ? reject(Error('Auth0 failed to sign up user')) : resolve()
    }))

    signUpPromise
      .then(() => {
        return Vue.prototype.$apiService.processNewSignup({
          company_name: companyName,
          company_website: companyWebsite,
          email: email,
          first_name: firstName,
          last_name: lastName
        })
      })
      .then(() => {
        // Once the user successfully signs up, use the username and password to log them in seemlessly
        this.login(email, password)
      })
      .catch((err: Error) => {
        console.log('Something went wrong during the sign up process')
        console.log(err)
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
    // Auth0 sucks so this is a hack to make the callback of check session resolve as a promise -> undefined if the user is logged out of the result of fetching the local token
    const authResult: Auth0DecodedHash = await new Promise((resolve: any, reject: any) => this.customClient.checkSession({}, (err: Auth0Error | null, result: Auth0DecodedHash) => {
      if (err) {
        resolve(undefined)
      } else {
        resolve(result)
      }
    }))

    if (authResult) {
      return store.dispatch('processAuthChange', authResult)
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

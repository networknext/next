import Vue from 'vue'
import APIService from '@/services/api.service'
import store from '@/store'

import { Auth0Client, IdToken } from '@auth0/auth0-spa-js'

export class AuthService {
  private apiService: APIService

  // in case we move the new Auth0Client() call out of the constructor
  private clientID: string
  private domain: string

  private authClient: any = null // Promise<Auth0Client>?
  popupOpen = false
  isAuthenticated = false
  user: any

  constructor (options: any) {
    this.apiService = Vue.prototype.$apiService
    this.clientID = options.clientId
    this.domain = options.domain

    this.authClient = new Auth0Client(({
      domain: this.domain,
      client_id: this.clientID,
      // audience: options.audience,
      redirect_uri: 'http://127.0.0.1:8080',
      advancedOptions: {
        defaultScope: 'openid profile email user_metadata app_metadata'
      },
      cacheLocation: 'localstorage'
    }))
  }

  async login (o: any) {
    this.popupOpen = true
    let idToken: any // IdToken // auth0 IdToken?

    try {
      await this.authClient.loginWithPopup()
    } catch (e) {
      console.log('login() error caught:')
      console.error(e)
    } finally {
      idToken = await this.authClient.getIdTokenClaims()
      console.log('idToken: ')
      console.log(idToken.__raw)
      this.popupOpen = false
    }

    this.user = await this.authClient.getUser()
    console.log('user: ')
    console.log(this.user)
    // console.log(this.authClient.user.name)

    // console.log('login()')
    // console.log(JSON.parse(idToken))
    this.processAuthentication(idToken)
    this.isAuthenticated = true
  }

  public processAuthentication (authResult: IdToken) {
    // console.log('processAuthentication() authResult:')
    // console.log(JSON.stringify(authResult))
    this.apiService = new APIService()
    const roles = authResult['https://networknext.com/userRoles'] || { roles: [] }
    const email = authResult.email || ''
    const domain = email.split('@')[1]
    const token = authResult.__raw
    const userProfile: UserProfile = {
      auth0ID: authResult.sub,
      company: '',
      email: authResult.email || '',
      idToken: token,
      name: authResult.name || '',
      roles: roles.roles,
      verified: authResult.email_verified || false,
      routeShader: null,
      domain: domain,
      pubKey: '',
      buyerID: ''
    }
    const promises = [
      this.apiService.fetchUserAccount({ user_id: userProfile.auth0ID }, token),
      this.apiService.fetchGameConfiguration({ domain: domain }, token),
      this.apiService.fetchAllBuyers(token)
    ]
    Promise.all(promises).then((responses: any) => {
      userProfile.buyerID = responses[0].account.buyer_id
      userProfile.company = responses[1].game_config.company
      userProfile.pubKey = responses[1].game_config.public_key
      userProfile.routeShader = responses[1].customer_route_shader
      const allBuyers = responses[2].buyers || []
      store.commit('UPDATE_USER_PROFILE', userProfile)
      store.commit('UPDATE_ALL_BUYERS', allBuyers)
    })

    // ToDo: we will need to pick on or the other
    // localStorage.setItem('userProfile', JSON.stringify(userProfile))
    localStorage.setItem('authResult', JSON.stringify(authResult))
  }
}

export const AuthPlugin = {
  install (Vue: any, options: any) {
    const client = new AuthService({
      domain: options.domain,
      clientId: options.clientId
    })

    Vue.mixin({
      created: function () {
        let err: any
        // console.log('created(), getting idtoken: ')
        try {
          client.processAuthentication(JSON.parse(localStorage.authResult))
          // console.log('created()')
          // console.log(JSON.parse(localStorage.authResult))
        } catch (err) {
          // console.log('processAuthentication(): nothing in localstorage')
        } finally {
          this.isAuthenticated = true
        }
      }
    })

    Vue.login = () => {
      console.log('login(): ' + options.domain)
      client.login(options)
    }

    Vue.logout = () => {
      console.log('logout(): ' + options.domain)
    }

    Vue.getUserInfo = () => {
      console.log('getUserInfo(): ' + options.domain)
    }
  }
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
  domain: string;
  pubKey: string;
  buyerID: string;
}

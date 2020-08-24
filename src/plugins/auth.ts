import Vue from 'vue'
import APIService from '@/services/api.service'
import store from '@/store'

import createAuth0Client, { Auth0Client, IdToken } from '@auth0/auth0-spa-js'

export class AuthService {
  private apiService: APIService
  private clientID: string
  private domain: string
  private authClient: Auth0Client | any

  constructor (options: any) {
    this.apiService = Vue.prototype.$apiService
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
    this.authClient.loginWithPopup()
      .then(() => {
        this.processAuthentication()
      })
      .catch((error: Error) => {
        console.log('login() error caught:')
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
    const isAuthenticated =
      await this.authClient.isAuthenticated()
        .catch((error: Error) => {
          console.log('something went wrong checking auth status')
          console.log(error)
        })

    if (!isAuthenticated) {
      return
    }
    this.apiService = new APIService()
    const userProfile: UserProfile = {
      auth0ID: '',
      company: '',
      email: '',
      idToken: '',
      name: '',
      roles: [],
      verified: false,
      routeShader: null,
      domain: '',
      pubKey: '',
      buyerID: ''
    }

    this.authClient.getIdTokenClaims().then((authResult: any) => {
      const roles: Array<any> = authResult['https://networknext.com/userRoles'].roles || { roles: [] }
      const email = authResult.email || ''
      const domain = email.split('@')[1]
      const token = authResult.__raw

      userProfile.roles = roles
      userProfile.domain = domain
      userProfile.email = email
      userProfile.idToken = token
      userProfile.auth0ID = authResult.sub

      return Promise.all([
        this.apiService.fetchUserAccount({ user_id: userProfile.auth0ID }, token),
        this.apiService.fetchGameConfiguration({ domain: domain }, token),
        this.apiService.fetchAllBuyers(token)
      ])
    }).then((responses: any) => {
      userProfile.buyerID = responses[0].account.buyer_id
      userProfile.company = responses[1].game_config.company
      userProfile.pubKey = responses[1].game_config.public_key
      userProfile.routeShader = responses[1].customer_route_shader
      const allBuyers = responses[2].buyers || []
      store.commit('UPDATE_USER_PROFILE', userProfile)
      store.commit('UPDATE_ALL_BUYERS', allBuyers)
    }).catch((error: Error) => {
      console.log('Something went wrong fetching user details')
      console.log(error.message)
    })
  }
}

export const AuthPlugin = {
  install (Vue: any, options: any) {
    let client: any = null
    Vue.mixin({
      created: () => {
        client = new AuthService({
          domain: options.domain,
          clientID: options.clientID
        })
      }
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

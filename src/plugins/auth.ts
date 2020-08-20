import Vue from 'vue'
import APIService from '@/services/api.service'
import { Auth0Client } from '@auth0/auth0-spa-js'

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
      redirect_uri: 'http://127.0.0.1:8080'
    }))
  }

  async login (o: any) {
    this.popupOpen = true
    let token: any
    let claims: any

    try {
      await this.authClient.loginWithPopup()
    } catch (e) {
      console.log('login() error caught:')
      console.error(e)
    } finally {
      token = await this.authClient.getTokenSilently()
      console.log('token: ')
      console.log(token)

      // claims = await this.authClient.getIdTokenClaims()
      // console.log('claims: ')
      // console.log(claims)
      // this.popupOpen = false
    }

    this.user = await this.authClient.getUser()
    console.log('user: ')
    console.log(this.user)
    // console.log(this.authClient.user.name)
    this.isAuthenticated = true
  }
}

export const AuthPlugin = {
  install (Vue: any, options: any) {
    const client = new AuthService({
      domain: options.domain,
      clientId: options.clientId
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

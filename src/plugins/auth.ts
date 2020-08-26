import Vue from 'vue'
import APIService from '@/services/api.service'
import store from '@/store'

import { Auth0Client } from '@auth0/auth0-spa-js'

import { UserProfile } from '@/components/types/AuthTypes.ts'

export class AuthService {
  private apiService: APIService
  private clientID: string
  private domain: string
  public authClient: Auth0Client | any

  constructor(options: any) {
    console.log('AuthService constructor()')
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

  public logout() {
    this.authClient.logout()
  }

  public login() {
    this.authClient
      .loginWithPopup()
      .then(() => {
        this.processAuthentication()
        console.log('login().then()')
      })
      .catch((error: Error) => {
        console.log('login() error caught:')
        console.error(error)
      })
  }

  public signUp() {
    this.authClient.loginWithPopup({
      connection: 'Username-Password-Authentication',
      screen_hint: 'signup'
    })
  }

  private async processAuthentication() {
    console.log('AuthService processAuthentication()')
    this.authClient
      .isAuthenticated()
      .then((isAuthenticated: boolean) => {
        if (!isAuthenticated) {
          console.log('!IsAuthenticated()')
          return
        }
        console.log('IsAuthenticated()')
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

        this.authClient
          .getIdTokenClaims()
          .then((authResult: any) => {
            const roles: Array<any> = authResult[
              'https://networknext.com/userRoles'
            ].roles || { roles: [] }
            const email = authResult.email || ''
            const domain = email.split('@')[1]
            const token = authResult.__raw

            userProfile.roles = roles
            userProfile.domain = domain
            userProfile.email = email
            userProfile.idToken = token
            userProfile.auth0ID = authResult.sub

            localStorage.setItem('userProfile', JSON.stringify(userProfile))
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
  install(Vue: any, options: any) {
    const client = new AuthService({
      domain: options.domain,
      clientID: options.clientID
    })

    // Vue.mixin({
    //   created: function () {
    //     console.log('AuthPlugin created()')
    //   }
    // })

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

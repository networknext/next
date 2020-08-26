import Vue from 'vue'
import store from '@/store'
import { UserProfile } from '@/components/types/AuthTypes.ts'
import _ from 'lodash'

export class JsonRpcService {
  private headers: any = null
  private unwatch: any

  constructor () {
    console.log('JsonRPCService constructor()')
    this.headers = {
      Accept: 'application/json',
      'Accept-Encoding': 'gzip',
      'Content-Type': 'application/json'
    }

    this.unwatch = store.watch(
      (state, getters) => getters.idToken,
      (newValue, oldValue) => {
        // it is enough to know that the token has changed - the value is
        // not relevant here
        console.log(`Updating from ${oldValue} to ${newValue}`)
        this.processAuthChange(newValue)
      }
    )
  }

  private processAuthChange (idToken: any): void {
    // let userProfile: UserProfile
    const userProfile = _.cloneDeep(store.getters.userProfile)
    Promise.all([
      this.fetchUserAccount({ user_id: userProfile.auth0ID }, idToken),
      this.fetchGameConfiguration({ domain: userProfile.domain }, idToken),
      this.fetchAllBuyers(idToken)
    ])
      .then((responses: any) => {
        userProfile.buyerID = responses[0].account.buyer_id
        userProfile.company = responses[1].game_config.company
        userProfile.pubKey = responses[1].game_config.public_key
        userProfile.routeShader = responses[1].customer_route_shader
        const allBuyers = responses[2].buyers || []
        store.commit('UPDATE_USER_PROFILE', userProfile)
        store.commit('UPDATE_ALL_BUYERS', allBuyers)
      })
      .catch((error: Error) => {
        console.log('Something went wrong fetching user details')
        console.log(error.message)
      })
  }

  private call (method: string, params: any, token: string): Promise<any> {
    if (!store.getters.isAnonymous || token) {
      this.headers.Authorization = `Bearer ${store.getters.idToken || token}`
    }
    return new Promise((resolve: any, reject: any) => {
      const options = params || {}
      const id = 'id'
      fetch(`${process.env.VUE_APP_API_URL}/rpc`, {
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({
          jsonrpc: '2.0',
          method,
          params: options,
          id
        })
      }).then((response: any) => {
        response.json().then((json: any) => {
          if (json.error) {
            reject(new Error(json.error))
          }
          resolve(json.result)
        })
      })
    })
  }

  public fetchTotalSessionCounts (args: any) {
    console.log('JsonRpcService.fetchTotalSessionCounts()')
    return this.call('BuyersService.TotalSessions', args, '')
  }

  public fetchMapSessions (args: any) {
    console.log('JsonRpcService.fetchMapSessions()')
    return this.call('BuyersService.SessionMap', args, '')
  }

  public fetchSessionDetails (args: any) {
    return this.call('BuyersService.SessionDetails', args, '')
  }

  public fetchTopSessions (args: any) {
    return this.call('BuyersService.TopSessions', args, '')
  }

  public fetchAllBuyers (token: string) {
    return this.call('BuyersService.Buyers', {}, token)
  }

  public fetchUserSessions (args: any) {
    return this.call('BuyersService.UserSessions', args, '')
  }

  public fetchAllRoles () {
    return this.call('AuthService.AllRoles', {}, '')
  }

  public fetchAllAccounts (args: any) {
    return this.call('AuthService.AllAccounts', args, '')
  }

  public updateUserRoles (args: any) {
    return this.call('AuthService.UpdateUserRoles', args, '')
  }

  public deleteUserAccount (args: any) {
    return this.call('AuthService.DeleteUserAccount', args, '')
  }

  public addNewUserAccounts (args: any) {
    return this.call('AuthService.AddUserAccount', args, '')
  }

  public fetchUserAccount (args: any, token: string) {
    return this.call('AuthService.UserAccount', args, token)
  }

  public fetchGameConfiguration (args: any, token: string) {
    return this.call('BuyersService.GameConfiguration', args, token)
  }

  public updateRouteShader (args: any) {
    return this.call('BuyersService.UpdateRouteShader', args, '')
  }

  public updateGameConfiguration (args: any) {
    return this.call('BuyersService.UpdateGameConfiguration', args, '')
  }

  public resendVerificationEmail (args: any) {
    return this.call('AuthService.ResendVerificationEmail', args, '')
  }
}

export const JsonRPCPlugin = {
  install (Vue: any, options: any) {
    // options?
    const client = new JsonRpcService()

    Vue.fetchTotalSessionCounts = (args: any) => {
      console.log('Vue.fetchTotalSessionCounts()')
      client.fetchTotalSessionCounts(args)
    }

    Vue.fetchMapSessions = (args: any) => {
      console.log('Vue.fetchMapSessions()')
      client.fetchMapSessions(args)
    }

    Vue.fetchSessionDetails = (args: any) => {
      client.fetchSessionDetails(args)
    }

    Vue.fetchTopSessions = (args: any) => {
      client.fetchTopSessions(args)
    }

    Vue.fetchAllBuyers = (args: any) => {
      client.fetchAllBuyers(args)
    }

    Vue.fetchUserSessions = (args: any) => {
      client.fetchUserSessions(args)
    }

    Vue.fetchAllRoles = () => {
      client.fetchAllRoles()
    }

    Vue.fetchAllAccounts = (args: any) => {
      client.fetchAllAccounts(args)
    }

    Vue.updateUserRoles = (args: any) => {
      client.updateUserRoles(args)
    }

    Vue.deleteUserAccount = (args: any) => {
      client.deleteUserAccount(args)
    }

    Vue.addNewUserAccounts = (args: any) => {
      client.addNewUserAccounts(args)
    }

    Vue.fetchUserAccount = (args: any, token: string) => {
      client.fetchUserAccount(args, token)
    }

    Vue.fetchGameConfiguration = (args: any, token: string) => {
      client.fetchGameConfiguration(args, token)
    }

    Vue.updateRouteShader = (args: any) => {
      client.updateRouteShader(args)
    }

    Vue.updateGameConfiguration = (args: any) => {
      client.updateGameConfiguration(args)
    }

    Vue.resendVerificationEmail = (args: any) => {
      client.resendVerificationEmail(args)
    }
  }
}

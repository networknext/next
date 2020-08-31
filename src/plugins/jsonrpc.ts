import store from '@/store'
import _ from 'lodash'

export class JsonRpcService {
  private headers: any

  constructor () {
    this.headers = {
      Accept: 'application/json',
      'Accept-Encoding': 'gzip',
      'Content-Type': 'application/json'
    }

    store.watch(
      (_, getters: any) => getters.idToken,
      () => {
        this.processAuthChange()
      }
    )
  }

  private processAuthChange (): void {
    const userProfile = _.cloneDeep(store.getters.userProfile)
    Promise.all([
      this.fetchUserAccount({ user_id: userProfile.auth0ID }, userProfile.token),
      this.fetchGameConfiguration({ domain: userProfile.domain }, userProfile.token),
      this.fetchAllBuyers(userProfile.token)
    ])
      .then((responses: any) => {
        userProfile.buyerID = responses[0].account.buyer_id
        userProfile.company = responses[1].game_config.company
        userProfile.pubKey = responses[1].game_config.public_key
        userProfile.routeShader = responses[1].customer_route_shader
        const allBuyers = responses[2].buyers || []
        store.commit('UPDATE_USER_PROFILE', userProfile)
        store.commit('UPDATE_ALL_BUYERS', allBuyers)
        // If the user has no roles, check to see if they are the only one in the company and upgrade their account to owner
        // TODO: Figure out a better way of doing this...
        if (!store.getters.isAnonymous && !store.getters.isAnonymousPlus && userProfile.roles.length === 0) {
          this.upgradeAccount({ user_id: userProfile.auth0ID }, userProfile.token)
            .then((response: any) => {
              const newRoles = response.new_roles || []
              if (newRoles.length > 0) {
                userProfile.roles = newRoles
              }
              store.commit('UPDATE_USER_PROFILE', userProfile)
            })
            .catch((error) => {
              console.log('Something went wrong upgrading the account')
              console.log(error)
            })
          return
        }
        store.commit('UPDATE_USER_PROFILE', userProfile)
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

  public upgradeAccount (args: any, token: string): Promise<any> {
    return this.call('AuthService.UpgradeAccount', args, token)
  }

  public fetchTotalSessionCounts (args: any): Promise<any> {
    return this.call('BuyersService.TotalSessions', args, '')
  }

  public fetchMapSessions (args: any): Promise<any> {
    return this.call('BuyersService.SessionMap', args, '')
  }

  public fetchSessionDetails (args: any): Promise<any> {
    return this.call('BuyersService.SessionDetails', args, '')
  }

  public fetchTopSessions (args: any): Promise<any> {
    return this.call('BuyersService.TopSessions', args, '')
  }

  public fetchAllBuyers (token: string): Promise<any> {
    return this.call('BuyersService.Buyers', {}, token)
  }

  public fetchUserSessions (args: any): Promise<any> {
    return this.call('BuyersService.UserSessions', args, '')
  }

  public fetchAllRoles (): Promise<any> {
    return this.call('AuthService.AllRoles', {}, '')
  }

  public fetchAllAccounts (args: any): Promise<any> {
    return this.call('AuthService.AllAccounts', args, '')
  }

  public updateUserRoles (args: any): Promise<any> {
    return this.call('AuthService.UpdateUserRoles', args, '')
  }

  public deleteUserAccount (args: any): Promise<any> {
    return this.call('AuthService.DeleteUserAccount', args, '')
  }

  public addNewUserAccounts (args: any): Promise<any> {
    return this.call('AuthService.AddUserAccount', args, '')
  }

  public fetchUserAccount (args: any, token: string): Promise<any> {
    return this.call('AuthService.UserAccount', args, token)
  }

  public fetchGameConfiguration (args: any, token: string): Promise<any> {
    return this.call('BuyersService.GameConfiguration', args, token)
  }

  public updateRouteShader (args: any): Promise<any> {
    return this.call('BuyersService.UpdateRouteShader', args, '')
  }

  public updateGameConfiguration (args: any): Promise<any> {
    return this.call('BuyersService.UpdateGameConfiguration', args, '')
  }

  public resendVerificationEmail (args: any): Promise<any> {
    return this.call('AuthService.ResendVerificationEmail', args, '')
  }
}

export const JsonRPCPlugin = {
  install (Vue: any) {
    const client = new JsonRpcService()

    Vue.fetchTotalSessionCounts = (args: any): Promise<any> => {
      return client.fetchTotalSessionCounts(args)
    }

    Vue.fetchMapSessions = (args: any): Promise<any> => {
      return client.fetchMapSessions(args)
    }

    Vue.fetchSessionDetails = (args: any): Promise<any> => {
      return client.fetchSessionDetails(args)
    }

    Vue.fetchTopSessions = (args: any): Promise<any> => {
      return client.fetchTopSessions(args)
    }

    Vue.fetchAllBuyers = (args: any): Promise<any> => {
      return client.fetchAllBuyers(args)
    }

    Vue.fetchUserSessions = (args: any): Promise<any> => {
      return client.fetchUserSessions(args)
    }

    Vue.fetchAllRoles = (): Promise<any> => {
      return client.fetchAllRoles()
    }

    Vue.fetchAllAccounts = (args: any): Promise<any> => {
      return client.fetchAllAccounts(args)
    }

    Vue.updateUserRoles = (args: any): Promise<any> => {
      return client.updateUserRoles(args)
    }

    Vue.deleteUserAccount = (args: any): Promise<any> => {
      return client.deleteUserAccount(args)
    }

    Vue.addNewUserAccounts = (args: any): Promise<any> => {
      return client.addNewUserAccounts(args)
    }

    Vue.fetchUserAccount = (args: any, token: string): Promise<any> => {
      return client.fetchUserAccount(args, token)
    }

    Vue.fetchGameConfiguration = (args: any, token: string): Promise<any> => {
      return client.fetchGameConfiguration(args, token)
    }

    Vue.updateRouteShader = (args: any): Promise<any> => {
      return client.updateRouteShader(args)
    }

    Vue.updateGameConfiguration = (args: any): Promise<any> => {
      return client.updateGameConfiguration(args)
    }

    Vue.resendVerificationEmail = (args: any): Promise<any> => {
      return client.resendVerificationEmail(args)
    }
  }
}

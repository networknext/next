import store from '@/store'
import _ from 'lodash'
import Vue from 'vue'

export class JSONRPCService {
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
      this.fetchUserAccount({ user_id: userProfile.auth0ID }),
      this.fetchGameConfiguration({ domain: userProfile.domain }),
      this.fetchAllBuyers()
    ])
      .then((responses: any) => {
        // userProfile.buyerID = responses[0].account.buyer_id
        userProfile.buyerID = responses[0].account.id
        userProfile.pubKey = responses[1].game_config.public_key
        // userProfile.routeShader = responses[1].customer_route_shader
        const allBuyers = responses[2].buyers || []
        store.commit('UPDATE_USER_PROFILE', userProfile)
        store.commit('UPDATE_ALL_BUYERS', allBuyers)
        // If the user has no roles, check to see if they are the only one in the company and upgrade their account to owner
        // TODO: Figure out a better way of doing this...
        if (!store.getters.isAnonymous && !store.getters.isAnonymousPlus && userProfile.roles.length === 0) {
          this.upgradeAccount({ user_id: userProfile.auth0ID })
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

  private call (method: string, params: any): Promise<any> {
    if (!store.getters.isAnonymous) {
      this.headers.Authorization = `Bearer ${store.getters.idToken}`
    }
    return new Promise((resolve: any, reject: any) => {
      const options = params || {}
      const id = 'id'
      let url = ''
      if (process.env.VUE_APP_MODE === 'local') {
        url = `${process.env.VUE_APP_API_URL}`
      }
      fetch(`${url}/rpc`, {
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

  public updateAccountSettings (args: any): Promise<any> {
    return this.call('AuthService.UpdateAccountSettings', args)
  }

  public updateCompanyName (args: any): Promise<any> {
    return this.call('AuthService.UpdateCompanyName', args)
  }

  public upgradeAccount (args: any): Promise<any> {
    return this.call('AuthService.UpgradeAccount', args)
  }

  public fetchTotalSessionCounts (args: any): Promise<any> {
    return this.call('BuyersService.TotalSessions', args)
  }

  public fetchMapSessions (args: any): Promise<any> {
    return this.call('BuyersService.SessionMap', args)
  }

  public fetchSessionDetails (args: any): Promise<any> {
    return this.call('BuyersService.SessionDetails', args)
  }

  public fetchTopSessions (args: any): Promise<any> {
    return this.call('BuyersService.TopSessions', args)
  }

  public fetchAllBuyers (): Promise<any> {
    return this.call('BuyersService.Buyers', {})
  }

  public fetchUserSessions (args: any): Promise<any> {
    return this.call('BuyersService.UserSessions', args)
  }

  public fetchAllRoles (): Promise<any> {
    return this.call('AuthService.AllRoles', {})
  }

  public fetchAllAccounts (): Promise<any> {
    return this.call('AuthService.AllAccounts', {})
  }

  public updateUserRoles (args: any): Promise<any> {
    return this.call('AuthService.UpdateUserRoles', args)
  }

  public deleteUserAccount (args: any): Promise<any> {
    return this.call('AuthService.DeleteUserAccount', args)
  }

  public addNewUserAccounts (args: any): Promise<any> {
    return this.call('AuthService.AddUserAccount', args)
  }

  public fetchUserAccount (args: any): Promise<any> {
    return this.call('AuthService.UserAccount', args)
  }

  public fetchGameConfiguration (args: any): Promise<any> {
    return this.call('BuyersService.GameConfiguration', args)
  }

  public updateRouteShader (args: any): Promise<any> {
    return this.call('BuyersService.UpdateRouteShader', args)
  }

  public updateGameConfiguration (args: any): Promise<any> {
    return this.call('BuyersService.UpdateGameConfiguration', args)
  }

  public resendVerificationEmail (args: any): Promise<any> {
    return this.call('AuthService.ResendVerificationEmail', args)
  }
}

export const JSONRPCPlugin = {
  service: {} as JSONRPCService,
  install (Vue: any) {
    this.service = new JSONRPCService()
    Vue.prototype.$apiService = this.service
  },
  fetchTotalSessionCounts: function (args: any): Promise<any> {
    return this.service.fetchTotalSessionCounts(args)
  },
  fetchMapSessions: function (args: any): Promise<any> {
    return this.service.fetchMapSessions(args)
  },
  fetchSessionDetails: function (args: any): Promise<any> {
    return this.service.fetchSessionDetails(args)
  },
  fetchTopSessions: function (args: any): Promise<any> {
    return this.service.fetchTopSessions(args)
  },
  fetchAllBuyers: function (): Promise<any> {
    return this.service.fetchAllBuyers()
  },
  fetchUserSessions: function (args: any): Promise<any> {
    return this.service.fetchUserSessions(args)
  },
  fetchAllRoles: function (): Promise<any> {
    return this.service.fetchAllRoles()
  },
  fetchAllAccounts: function (): Promise<any> {
    return this.service.fetchAllAccounts()
  },
  updateUserRoles: function (args: any): Promise<any> {
    return this.service.updateUserRoles(args)
  },
  deleteUserAccount: function (args: any): Promise<any> {
    return this.service.deleteUserAccount(args)
  },
  addNewUserAccounts: function (args: any): Promise<any> {
    return this.service.addNewUserAccounts(args)
  },
  fetchUserAccount: function (args: any): Promise<any> {
    return this.service.fetchUserAccount(args)
  },
  fetchGameConfiguration: function (args: any): Promise<any> {
    return this.service.fetchGameConfiguration(args)
  },
  updateRouteShader: function (args: any): Promise<any> {
    return this.service.updateRouteShader(args)
  },
  updateGameConfiguration: function (args: any): Promise<any> {
    return this.service.updateGameConfiguration(args)
  },
  resendVerificationEmail: function (args: any): Promise<any> {
    return this.service.resendVerificationEmail(args)
  }
}

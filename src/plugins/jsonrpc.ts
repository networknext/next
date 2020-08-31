import store from '@/store'
import _ from 'lodash'

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
      (newValue: string) => {
        // it is enough to know that the token has changed - the value is
        // not relevant here
        this.processAuthChange(newValue)
      }
    )
  }

  private processAuthChange (idToken: string): void {
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

export const JSONRPCPlugin = {
  service: {} as JSONRPCService,
  install (Vue: any) {
    this.service = new JSONRPCService()

    Vue.fetchTotalSessionCounts = (args: any): Promise<any> => {
      return this.service.fetchTotalSessionCounts(args)
    }

    Vue.fetchMapSessions = (args: any): Promise<any> => {
      return this.service.fetchMapSessions(args)
    }

    Vue.fetchSessionDetails = (args: any): Promise<any> => {
      return this.service.fetchSessionDetails(args)
    }

    Vue.fetchTopSessions = (args: any): Promise<any> => {
      return this.service.fetchTopSessions(args)
    }

    Vue.fetchAllBuyers = (args: any): Promise<any> => {
      return this.service.fetchAllBuyers(args)
    }

    Vue.fetchUserSessions = (args: any): Promise<any> => {
      return this.service.fetchUserSessions(args)
    }

    Vue.fetchAllRoles = (): Promise<any> => {
      return this.service.fetchAllRoles()
    }

    Vue.fetchAllAccounts = (args: any): Promise<any> => {
      return this.service.fetchAllAccounts(args)
    }

    Vue.updateUserRoles = (args: any): Promise<any> => {
      return this.service.updateUserRoles(args)
    }

    Vue.deleteUserAccount = (args: any): Promise<any> => {
      return this.service.deleteUserAccount(args)
    }

    Vue.addNewUserAccounts = (args: any): Promise<any> => {
      return this.service.addNewUserAccounts(args)
    }

    Vue.fetchUserAccount = (args: any, token: string): Promise<any> => {
      return this.service.fetchUserAccount(args, token)
    }

    Vue.fetchGameConfiguration = (args: any, token: string): Promise<any> => {
      return this.service.fetchGameConfiguration(args, token)
    }

    Vue.updateRouteShader = (args: any): Promise<any> => {
      return this.service.updateRouteShader(args)
    }

    Vue.updateGameConfiguration = (args: any): Promise<any> => {
      return this.service.updateGameConfiguration(args)
    }

    Vue.resendVerificationEmail = (args: any): Promise<any> => {
      return this.service.resendVerificationEmail(args)
    }
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
  fetchAllBuyers: function (args: any): Promise<any> {
    return this.service.fetchAllBuyers(args)
  },
  fetchUserSessions: function (args: any): Promise<any> {
    return this.service.fetchUserSessions(args)
  },
  fetchAllRoles: function (): Promise<any> {
    return this.service.fetchAllRoles()
  },
  fetchAllAccounts: function (args: any): Promise<any> {
    return this.service.fetchAllAccounts(args)
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
  fetchUserAccount: function (args: any, token: string): Promise<any> {
    return this.service.fetchUserAccount(args, token)
  },
  fetchGameConfiguration: function (args: any, token: string): Promise<any> {
    return this.service.fetchGameConfiguration(args, token)
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

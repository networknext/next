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
      () => {
        this.processAuthChange()
      }
    )
  }

  private processAuthChange (): void {
    const userProfile = _.cloneDeep(store.getters.userProfile)
    let promises = []
    if (store.getters.registeredToCompany) {
      promises = [
        this.fetchUserAccount({ user_id: userProfile.auth0ID }),
        this.fetchGameConfiguration(),
        this.fetchAllBuyers()
      ]
    } else {
      promises = [
        this.fetchUserAccount({ user_id: userProfile.auth0ID }),
        this.fetchAllBuyers()
      ]
    }
    Promise.all(promises)
      .then((responses: any) => {
        let allBuyers = []
        if (store.getters.registeredToCompany) {
          allBuyers = responses[2].buyers
          userProfile.pubKey = responses[1].game_config.public_key
        } else {
          allBuyers = responses[1].buyers
        }
        userProfile.buyerID = responses[0].account.id
        userProfile.companyName = responses[0].account.company_name || ''
        userProfile.domains = responses[0].domains || []
        store.commit('UPDATE_USER_PROFILE', userProfile)
        store.commit('UPDATE_ALL_BUYERS', allBuyers)
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

  public impersonate (args: any): Promise<any> {
    return this.call('AuthService.Impersonate', args)
  }

  public updateAccountSettings (args: any): Promise<any> {
    return this.call('AuthService.UpdateAccountSettings', args)
  }

  public updateAutoSignupDomains (args: any): Promise<any> {
    return this.call('AuthService.UpdateAutoSignupDomains', args)
  }

  public updateCompanyInformation (args: any): Promise<any> {
    return this.call('AuthService.UpdateCompanyInformation', args)
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

  public fetchGameConfiguration (): Promise<any> {
    return this.call('BuyersService.GameConfiguration', {})
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
  }
}

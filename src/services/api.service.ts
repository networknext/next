import store from '@/store'

export default class APIService {
  private headers: any = null;

  constructor () {
    this.headers = {
      Accept: 'application/json',
      'Accept-Encoding': 'gzip',
      'Content-Type': 'application/json'
    }
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
    return this.call('BuyersService.TotalSessions', args, '')
  }

  public fetchMapSessions (args: any) {
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

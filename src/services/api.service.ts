import store from '@/store'

/**
 * Rough attempt at making a "service" in Vue. It should be modeling how Angular handles services
 *  but definitely isn't perfect. This service handles all of the different API calls to the JSONRPC backend.
 *  Function calls are defined here rather than in their associated components because this allows us to stub
 *  them in tests.
 */

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
    console.log('jsonrpc method: ' + method)
    if (!store.getters.isAnonymous || token) {
      this.headers.Authorization = `Bearer ${store.getters.idToken || token}`
    }
    return new Promise((resolve: any, reject: any) => {
      const options = params || {}
      const id = 'id'
      const jsonrpc = 'http://172.23.165.40:20000'
      // fetch(`${process.env.VUE_APP_API_URL}/rpc`, {
      fetch(`${jsonrpc}/rpc`, {
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
      }).catch((err: any) => {
        console.log('fetch() error:')
        console.log(err)
      })
    })
  }

  // TODO: It may be better to make a generic call that takes the string name of the endpoint and the args for the endpoint
  //        that way the calls are still stubable and we can keep each call in the associated component. IE: SessionMap call
  //        in the SessionMap component...
  //        Yeah I like that a lot better

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

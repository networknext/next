import store from '@/store'

export class JSONRPCService {
  private headers: any

  constructor () {
    this.headers = {
      Accept: 'application/json',
      'Accept-Encoding': 'gzip',
      'Content-Type': 'application/json'
    }
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
      }).then((response: Response) => {
        return response.json()
      }).then((json: any) => {
        if (json.error) {
          reject(json.error)
          return
        }
        resolve(json.result)
      }).catch((error: Error) => {
        reject(error)
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

  public setupCompanyAccount (args: any): Promise<any> {
    return this.call('AuthService.SetupCompanyAccount', args)
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

  public fetchAnalyticsDashboards (args: any): Promise<any> {
    return this.call('BuyersService.FetchAnalyticsDashboards', args)
  }

  public fetchUsageSummary (args: any): Promise<any> {
    return this.call('BuyersService.FetchUsageDashboard', args)
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

  public updateAccountDetails (args: any): Promise<any> {
    return this.call('AuthService.UpdateAccountDetails', args)
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

  public fetchFeatureFlags (): Promise<any> {
    return this.call('ConfigService.AllFeatureFlags', {})
  }

  public sendDocsViewSlackNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerViewedTheDocsSlackNotification', args)
  }

  public sendResetPasswordEmail (args: any): Promise<any> {
    return this.call('AuthService.ResetPasswordEmail', args)
  }

  public sendSDKDownloadSlackNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerDownloadedSDKSlackNotification', args)
  }

  public sendPublicKeyEnteredSlackNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerEnteredPublicKeySlackNotification', args)
  }

  public sendUE4DownloadNotifications (args: any): Promise<any> {
    return this.call('AuthService.CustomerDownloadedUE4PluginNotifications', args)
  }

  public startAnalyticsTrial (): Promise<any> {
    return this.call('BuyersService.StartAnalyticsTrial', {})
  }

  public fetchNotifications (): Promise<any> {
    return this.call('BuyersService.FetchNotifications', {})
  }

  public fetchLookerURL (): Promise<any> {
    return this.call('BuyersService.FetchLookerURL', {})
  }

  public processNewSignup (args: any): Promise<any> {
    return this.call('AuthService.ProcessNewSignup', args)
  }

  public fetchDiscoveryDashboards (args: any): Promise<any> {
    return this.call('BuyersService.FetchDiscoveryDashboards', args)
  }
}

export const JSONRPCPlugin = {
  install (Vue: any) {
    Vue.$apiService = Vue.prototype.$apiService = new JSONRPCService()
  }
}

import store from '@/store'

export class JSONRPCService {
  private headers: any
  private url: string

  constructor () {
    this.headers = {
      Accept: 'application/json',
      'Accept-Encoding': 'gzip',
      'Content-Type': 'application/json'
    }
    this.url = process.env.VUE_APP_MODE === 'local' ? `${process.env.VUE_APP_API_URL}` : ''
  }

  private internalCall (endpoint: string): Promise<any> {
    return new Promise((resolve: any, reject: any) => {
      fetch(`${this.url}/${endpoint}`, {
        headers: {
          Accept: 'application/json',
          'Accept-Encoding': 'gzip',
          'Content-Type': 'application/json'
        },
        method: 'POST'
      }).then((response: Response) => {
        return response.json()
      }).then((json: any) => {
        if (json.error) {
          reject(json.error)
          return
        }
        resolve(json)
      }).catch((error: Error) => {
        reject(error)
      })
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

  public fetchPortalVersion (): Promise<any> {
    return this.internalCall('version')
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

  public fetchDiscoveryDashboards (args: any): Promise<any> {
    return this.call('BuyersService.FetchDiscoveryDashboards', args)
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

  public fetchCurrentSaves (args: any): Promise<any> {
    return this.call('BuyersService.FetchCurrentSaves', args)
  }

  public fetchSavesDashboard (args: any): Promise<any> {
    return this.call('BuyersService.FetchSavesDashboard', args)
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

  public sendSDKSourceViewSlackNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerViewedSDKSourceNotification', args)
  }

  public sendPublicKeyEnteredSlackNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerEnteredPublicKeySlackNotification', args)
  }

  public sendUE4DownloadNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerDownloadedUE4PluginNotification', args)
  }

  public sendUE4SourceViewNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerViewedUE4SourceNotification', args)
  }

  public send2022WhitePaperDownloadNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerDownloaded2022WhitePaperNotification', args)
  }

  public sendENetDownloadNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerDownloadedENetDownloadNotification', args)
  }

  public sendENetSourceViewSlackNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerViewedENetSourceNotification', args)
  }

  public sendUnityDownloadNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerDownloadedUnityPluginNotification', args)
  }

  public sendUnitySourceViewNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerViewedUnitySourceNotification', args)
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

  public buyerTOSSigned (args: any): Promise<any> {
    return this.call('BuyersService.SignedBuyerTOS', args)
  }
}

export const JSONRPCPlugin = {
  install (Vue: any) {
    Vue.$apiService = Vue.prototype.$apiService = new JSONRPCService()
  }
}

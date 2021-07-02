import { FeatureEnum } from '@/components/types/FeatureTypes'
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
    return this.call('BuyersService.UpdateBuyerInformation', args)
  }

  public resendVerificationEmail (args: any): Promise<any> {
    return this.call('AuthService.ResendVerificationEmail', args)
  }

  public fetchFeatureFlags (): Promise<any> {
    return this.call('ConfigService.AllFeatureFlags', {})
  }

  public sendSignUpSlackNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerSignedUpSlackNotification', args)
  }

  public sendDocsViewSlackNotification (args: any): Promise<any> {
    return this.call('AuthService.CustomerViewedTheDocsSlackNotification', args)
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

  public fetchNotifications (): Promise<any> {
    return this.call('BuyersService.FetchNotifications', {})
  }

  public fetchLookerURL (): Promise<any> {
    return this.call('BuyersService.FetchLookerURL', {})
  }
}

export const JSONRPCPlugin = {
  install (Vue: any) {
    Vue.$apiService = Vue.prototype.$apiService = new JSONRPCService()
  }
}

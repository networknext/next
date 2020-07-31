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

  private call (method: string, params: any): Promise<any> {
    if (!store.getters.isAnonymous) {
      this.headers.Authorization = `Bearer ${store.getters.idToken}`
    }
    return new Promise((resolve, reject) => {
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
      }).then((response: Response) => {
        resolve(response.json())
      })
        .catch((error: Error) => {
          console.log(error.message)
          reject(error)
        })
    })
  }

  public fetchTotalSessionCounts (args: any) {
    return this.call('BuyersService.TotalSessions', args)
  }

  public fetchMapSessions (args: any) {
    return this.call('BuyersService.SessionMap', args)
  }

  public fetchSessionDetails (args: any) {
    return this.call('BuyersService.SessionDetails', args)
  }

  public fetchTopSessions (args: any) {
    return this.call('BuyersService.TopSessions', args)
  }
}

import store from '@/store'

export default class APIService {
  private headers: any = null;

  constructor () {
    this.headers = {
      Accept: 'application/jsosn',
      'Accept-Encoding': 'gzip',
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*'
    }
  }

  public call (method: string, params: any): Promise<any> {
    if (!store.getters.isAnonymous) {
      this.headers.Authorization = `Bearer ${store.getters.idToken}`
    }
    return new Promise((resolve, reject) => {
      const options = params || {}
      const id = 'id'
      fetch('/rpc', {
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({
          jsonrpc: '2.0',
          method,
          params: options,
          id
        })
      })
        .then((response: Response) => {
          resolve(response.json())
        })
        .catch((error: Error) => {
          reject(error)
        })
    })
  }
}

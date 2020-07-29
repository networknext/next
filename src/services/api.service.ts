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

  public call (method: string, params: any): Promise<any> {
    if (!store.getters.isAnonymous) {
      this.headers.Authorization = `Bearer ${store.getters.idToken}`
    }
    return new Promise((resolve, reject) => {
      const options = params || {}
      const id = 'id'
      const test = window.fetch(`${process.env.VUE_APP_BASE_URL}/rpc`, {
        method: 'POST',
        headers: this.headers,
        body: JSON.stringify({
          jsonrpc: '2.0',
          method,
          params: options,
          id
        })
      })
      console.log(test)
      test.then((response: Response) => {
        resolve(response.json())
      })
        .catch((error: Error) => {
          console.log(error.message)
          reject(error)
        })
    })
  }
}

export default class APIService {
  private headers: any = null;

  constructor () {
    this.headers = {
      Accept: 'application/jsosn',
      'Accept-Encoding': 'gzip',
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*'
    }

    /**
     * TODO: Check if there is an auth token associated with
     *       the user and add the token to the auth header
     *
     * */
  }

  public call (method: string, params: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const options = params || {}
      const id = JSON.stringify(params)
      fetch('http://127.0.0.1:20000/rpc', {
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

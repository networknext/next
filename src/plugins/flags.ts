import { JSONRPCService } from './jsonrpc'

export class FeatureFlagService {
  private apiService: JSONRPCService | null
  private flags: Array<any>

  constructor (options: any) {
    this.apiService = null
    if (options.useAPI) {
      this.apiService = options.apiService
    }
    this.flags = options.flags
  }

  private fetchAllRemoteFeatureFlags (): void {
    if (!this.apiService) {
      throw new Error('API Service not defined')
    }
    this.apiService.fetchFeatureFlags()
      .then((response: any) => {
        const newFlags: Array<any> = response.flags
        this.flags.forEach((flag: any) => {
          flag.value = newFlags[flag.name].value || false
        })
      })
  }

  private fetchEnvVarFeatureFlags () {
    this.flags.forEach((flag: any) => {
      const envVarString = `VUE_APP_${flag.name}`
      flag.value = process.env[envVarString] || false
    })
  }

  private isEnabled (name: string): boolean {
    this.flags.forEach((flag: any) => {
      if (flag.name === name) {
        return flag.value.toLowerCase() === 'true' || false
      }
    })
    return false
  }
}

export const FlagPlugin = {
  install (Vue: any, options: any) {
    Vue.$flagService = Vue.prototype.$flagService = new FeatureFlagService(options)
  }
}

import { FeatureEnum, Flag } from '@/components/types/FeatureTypes'
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
          flag.value = newFlags[flag.name] || false
        })
      })
  }

  private fetchEnvVarFeatureFlags () {
    this.flags.forEach((flag: Flag) => {
      const envVarString = `VUE_APP_${flag.name}`
      flag.value = process.env[envVarString] ? process.env[envVarString].toLowerCase() === 'true' : false
    })
  }

  private isEnabled (name: FeatureEnum): boolean {
    let value = false
    this.flags.forEach((flag: any) => {
      if (flag.name === name) {
        switch (typeof flag.value) {
          case 'boolean':
            value = flag.value
            break
          default:
            throw new Error('Unknown flag')
        }
      }
    })
    return value
  }

  private printEnabledFlags (): void {
    this.flags.forEach((flag: Flag) => {
      if (flag.value) {
        console.log(flag)
      }
    })
  }
}

export const FlagPlugin = {
  install (Vue: any, options: any) {
    Vue.$flagService = Vue.prototype.$flagService = new FeatureFlagService(options)
  }
}

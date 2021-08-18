declare module '*.vue' {
  import _Vue from 'vue'

  export default _Vue

  export interface VueJSONRPCService {
    addMailChimpContact (args: any): Promise<any>;
    impersonate (args: any): Promise<any>;
    updateAccountSettings (args: any): Promise<any>;
    updateAutoSignupDomains (args: any): Promise<any>;
    createCompanyAccount (args: any): Promise<any>;
    upgradeAccount (args: any): Promise<any>;
    fetchTotalSessionCounts (args: any): Promise<any>;
    fetchMapSessions (args: any): Promise<any>;
    fetchSessionDetails (args: any): Promise<any>;
    fetchTopSessions (args: any): Promise<any>;
    fetchAllBuyers (): Promise<any>;
    fetchUserSessions (args: any): Promise<any>;
    fetchAllRoles (): Promise<any>;
    fetchAllAccounts (): Promise<any>;
    updateAccountDetails (args: any): Promise<any>;
    updateUserRoles (args: any): Promise<any>;
    deleteUserAccount (args: any): Promise<any>;
    addNewUserAccounts (args: any): Promise<any>;
    fetchUserAccount (args: any): Promise<any>;
    fetchNotifications (): Promise<any>;
    fetchLookerURL (): Promise<any>;
    fetchGameConfiguration (): Promise<any>;
    updateRouteShader (args: any): Promise<any>;
    updateGameConfiguration (args: any): Promise<any>;
    resendVerificationEmail (args: any): Promise<any>;
    sendSignUpSlackNotification (args: any): Promise<any>;
    sendDocsViewSlackNotification (args: any): Promise<any>;
    sendSDKDownloadSlackNotification (args: any): Promise<any>;
    sendPublicKeyEnteredSlackNotification (args: any): Promise<any>;
    sendUE4DownloadNotifications (args: any): Promise<any>;
  }

  export class VueJSONRPCServicePlugin {
    static install(
      Vue: typeof _Vue,
    ): void
  }

  export interface VueAuthService {
    logout (): void;
    login (): void;
    signUp (): void;
    refreshToken (): Promise<any>;
  }

  export class VueAuthServicePlugin {
    static install(
      Vue: typeof _Vue,
    ): void
  }

  export interface VueFlagService {
    fetchEnvVarFeatureFlags (): Promise<any>;
    fetchAllRemoteFeatureFlags (): void;
    isEnabled (name: string): boolean;
  }

  export class VueFlagServicePlugin {
    static install(
      Vue: typeof _Vue,
      options: any,
    ): void
  }

  module 'vue/types/vue' {
    interface Vue {
      $apiService: VueJSONRPCService;
      $authService: VueAuthService;
      $flagService: VueFlagService;
    }
  }
}

declare module '*.vue' {
  import _Vue from 'vue'

  export default _Vue

  export interface VueJSONRPCService {
    addMailChimpContact (args: any): Promise<any>;
    impersonate (args: any): Promise<any>;
    updateAccountSettings (args: any): Promise<any>;
    updateAutoSignupDomains (args: any): Promise<any>;
    setupCompanyAccount (args: any): Promise<any>;
    upgradeAccount (args: any): Promise<any>;
    fetchPortalVersion (): Promise<any>;
    fetchTotalSessionCounts (args: any): Promise<any>;
    fetchMapSessions (args: any): Promise<any>;
    fetchSessionDetails (args: any): Promise<any>;
    fetchTopSessions (args: any): Promise<any>;
    fetchAllBuyers (): Promise<any>;
    fetchUsageSummary(args: any): Promise<any>;
    fetchAnalyticsDashboards(args: any): Promise<any>;
    fetchDiscoveryDashboards (args: any): Promise<any>;
    fetchSavesDashboard (args: any): Promise<any>;
    fetchUserSessions (args: any): Promise<any>;
    fetchAllRoles (): Promise<any>;
    fetchAllAccounts (): Promise<any>;
    fetchCurrentSaves (args: any): Promise<any>;
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
    sendDocsViewSlackNotification (args: any): Promise<any>;
    sendResetPasswordEmail (args: any): Promise<any>;
    sendSDKDownloadSlackNotification (args: any): Promise<any>;
    sendPublicKeyEnteredSlackNotification (args: any): Promise<any>;
    sendUE4DownloadNotifications (args: any): Promise<any>;
    startAnalyticsTrial (): Promise<any>;
    processNewSignup (args: any): Promise<any>;
  }

  export class VueJSONRPCServicePlugin {
    static install(
      Vue: typeof _Vue,
    ): void
  }

  export interface VueAuthService {
    login (username: string, password: string, redirectURI: string): Promise<any>;
    logout (): void;
    getAccess(email: string, password: string): Promise<Error | undefined>;
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

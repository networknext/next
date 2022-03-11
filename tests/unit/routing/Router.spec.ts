import { newDefaultProfile } from '@/components/types/AuthTypes'
import { FlagPlugin } from '@/plugins/flags'
import router from '@/router'
import store from '@/store'
import Vue from 'vue'
import VueRouter from 'vue-router'
import Vuex from 'vuex'

Vue.use(VueRouter)
Vue.use(FlagPlugin, {
  flags: [],
  useAPI: false,
  apiService: {}
})
Vue.use(Vuex)

describe('Router config', () => {
  it('should be in history mode', () => {
    expect(router.mode.toString()).toEqual('history')
  })
})

describe('Router Rules', () => {
  afterEach(() => {
    if (router.currentRoute.fullPath !== '/map') {
      router.push('/map')
    }
  })

  // TODO: This needs to be worked out better because the test still passes even though the expect fails - it throws an error in the console though ?!
  function assertNoErrorRouting(route: string) {
    router.push(route).catch((error: Error) => {
      expect(error.message).not.toContain('Route does not exist')
      expect(error.message).not.toContain('Insufficient privileges')
    })
  }

  function assertErrorRouting(route: string) {
    router.push(route).catch((error: Error) => {
      expect(error).not.toBeUndefined()
    })
  }

  it('checks anonymous navigation', async () => {
    expect(store.getters.isAnonymous).toBeTruthy()

    // Allowed Routes
    assertNoErrorRouting('/')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')
    assertNoErrorRouting('/user-tool')

    // Illegal Routes
    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fuser-tool')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fuser-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fuser-tool%2F00000000')

    assertNoErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fdownloads')

    assertNoErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fsettings')

    assertNoErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fsettings%2Faccount')

    assertNoErrorRouting('/settings/game-config')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fsettings%2Fgame-config')

    assertNoErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fsettings%2Fusers')

    assertNoErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fusage')

    assertNoErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fusage%2F2021-08')

    assertNoErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/login?redirectURI=%2Fanalytics')
  })

  it('checks anonymous plus navigation', async () => {
    const defaultUserProfile = newDefaultProfile()
    defaultUserProfile.idToken = 'not an empty token'

    store.dispatch('updateUserProfile', defaultUserProfile)

    expect(store.getters.isAnonymous).toBeFalsy()
    expect(store.getters.userProfile.verified).toBeFalsy()
    expect(store.getters.isAnonymousPlus).toBeTruthy()

    // Allowed Routes
    expect(router.currentRoute.fullPath).toEqual('/map')

    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/user-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/user-tool/00000000')

    // Illegal Routes

    assertErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/map')
  })

  it('checks viewer navigation', async () => {
    const defaultUserProfile = newDefaultProfile()
    defaultUserProfile.idToken = 'not an empty token'
    defaultUserProfile.verified = true

    store.dispatch('updateUserProfile', defaultUserProfile)

    expect(store.getters.isAnonymous).toBeFalsy()
    expect(store.getters.isAnonymousPlus).toBeFalsy()
    expect(store.getters.registeredToCompany).toBeFalsy()
    expect(store.getters.userProfile.verified).toBeTruthy()

    // Allowed Routes
    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/user-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/user-tool/00000000')

    assertNoErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertNoErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/downloads')

    assertNoErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    // Illegal Routes
    assertErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/game-config')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/map')
  })

  it('checks owner navigation', async () => {
    const defaultUserProfile = newDefaultProfile()
    defaultUserProfile.idToken = 'not an empty token'
    defaultUserProfile.verified = true
    defaultUserProfile.companyCode = 'test'
    defaultUserProfile.companyName = 'Test Company'
    defaultUserProfile.roles.push('Owner')

    store.dispatch('updateUserProfile', defaultUserProfile)

    expect(store.getters.isAnonymous).toBeFalsy()
    expect(store.getters.isAnonymousPlus).toBeFalsy()
    expect(store.getters.userProfile.verified).toBeTruthy()
    expect(store.getters.registeredToCompany).toBeTruthy()
    expect(store.getters.isOwner).toBeTruthy()

    // Allowed Routes
    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/user-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/user-tool/00000000')

    assertNoErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertNoErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/downloads')

    assertNoErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertErrorRouting('/settings/game-config')

    expect(router.currentRoute.fullPath).toEqual('/settings/game-config')

    assertErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/settings/users')

    // Illegal Routes
    assertErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/map')
  })

  it('checks explore - no billing or analytics', async () => {
    const defaultUserProfile = newDefaultProfile()
    defaultUserProfile.idToken = 'not an empty token'
    defaultUserProfile.verified = true
    defaultUserProfile.companyCode = 'test'
    defaultUserProfile.companyName = 'Test Company'
    defaultUserProfile.roles.push('Explorer')

    store.dispatch('updateUserProfile', defaultUserProfile)

    expect(store.getters.isAnonymous).toBeFalsy()
    expect(store.getters.isAnonymousPlus).toBeFalsy()
    expect(store.getters.userProfile.verified).toBeTruthy()
    expect(store.getters.registeredToCompany).toBeTruthy()
    expect(store.getters.isExplorer).toBeTruthy()

    // Allowed Routes
    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/user-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/user-tool/00000000')

    assertNoErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertNoErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/downloads')

    assertNoErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    // Illegal Routes
    assertErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/game-config')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/map')
  })

  it('checks explore - billing - no analytics', async () => {
    const defaultUserProfile = newDefaultProfile()
    defaultUserProfile.idToken = 'not an empty token'
    defaultUserProfile.verified = true
    defaultUserProfile.companyCode = 'test'
    defaultUserProfile.companyName = 'Test Company'
    defaultUserProfile.roles.push('Explorer')
    defaultUserProfile.hasBilling = true

    store.dispatch('updateUserProfile', defaultUserProfile)

    expect(store.getters.isAnonymous).toBeFalsy()
    expect(store.getters.isAnonymousPlus).toBeFalsy()
    expect(store.getters.userProfile.verified).toBeTruthy()
    expect(store.getters.registeredToCompany).toBeTruthy()
    expect(store.getters.isExplorer).toBeTruthy()

    // Allowed Routes
    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/user-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/user-tool/00000000')

    assertNoErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertNoErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/downloads')

    assertNoErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/usage')

    assertErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/usage/2021-08')

    // Illegal Routes

    assertErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/game-config')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/map')
  })

  it('checks explore - no billing - analytics', async () => {
    const defaultUserProfile = newDefaultProfile()
    defaultUserProfile.idToken = 'not an empty token'
    defaultUserProfile.verified = true
    defaultUserProfile.companyCode = 'test'
    defaultUserProfile.companyName = 'Test Company'
    defaultUserProfile.roles.push('Explorer')
    defaultUserProfile.hasAnalytics = true

    store.dispatch('updateUserProfile', defaultUserProfile)

    expect(store.getters.isAnonymous).toBeFalsy()
    expect(store.getters.isAnonymousPlus).toBeFalsy()
    expect(store.getters.userProfile.verified).toBeTruthy()
    expect(store.getters.registeredToCompany).toBeTruthy()
    expect(store.getters.isExplorer).toBeTruthy()

    // Allowed Routes
    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/user-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/user-tool/00000000')

    assertNoErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertNoErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/downloads')

    assertNoErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/analytics')

    // Illegal Routes

    assertErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/game-config')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/map')
  })

  it('checks explore - billing - analytics', async () => {
    const defaultUserProfile = newDefaultProfile()
    defaultUserProfile.idToken = 'not an empty token'
    defaultUserProfile.verified = true
    defaultUserProfile.companyCode = 'test'
    defaultUserProfile.companyName = 'Test Company'
    defaultUserProfile.roles.push('Explorer')
    defaultUserProfile.hasBilling = true
    defaultUserProfile.hasAnalytics = true

    store.dispatch('updateUserProfile', defaultUserProfile)

    expect(store.getters.isAnonymous).toBeFalsy()
    expect(store.getters.isAnonymousPlus).toBeFalsy()
    expect(store.getters.userProfile.verified).toBeTruthy()
    expect(store.getters.registeredToCompany).toBeTruthy()
    expect(store.getters.isExplorer).toBeTruthy()

    // Allowed Routes
    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/user-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/user-tool/00000000')

    assertNoErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertNoErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/downloads')

    assertNoErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/analytics')

    assertErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/usage')

    assertErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/usage/2021-08')

    // Illegal Routes

    assertErrorRouting('/settings/game-config')

    expect(router.currentRoute.fullPath).toEqual('/map')

    assertErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/map')
  })

  it('checks explore - billing - analytics', async () => {
    const defaultUserProfile = newDefaultProfile()
    defaultUserProfile.idToken = 'not an empty token'
    defaultUserProfile.verified = true
    defaultUserProfile.companyCode = 'test'
    defaultUserProfile.companyName = 'Test Company'
    defaultUserProfile.roles.push('Admin')

    store.dispatch('updateUserProfile', defaultUserProfile)

    expect(store.getters.isAnonymous).toBeFalsy()
    expect(store.getters.isAnonymousPlus).toBeFalsy()
    expect(store.getters.userProfile.verified).toBeTruthy()
    expect(store.getters.registeredToCompany).toBeTruthy()
    expect(store.getters.isAdmin).toBeTruthy()

    // Allowed Routes
    assertNoErrorRouting('/sessions')

    expect(router.currentRoute.fullPath).toEqual('/sessions')

    assertNoErrorRouting('/session-tool')

    expect(router.currentRoute.fullPath).toEqual('/session-tool')

    assertNoErrorRouting('/session-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/session-tool/00000000')

    assertNoErrorRouting('/explore/saves')

    expect(router.currentRoute.fullPath).toEqual('/explore/saves')

    assertNoErrorRouting('/user-tool')

    expect(router.currentRoute.fullPath).toEqual('/user-tool')

    assertNoErrorRouting('/user-tool/00000000')

    expect(router.currentRoute.fullPath).toEqual('/user-tool/00000000')

    assertNoErrorRouting('/settings')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertNoErrorRouting('/downloads')

    expect(router.currentRoute.fullPath).toEqual('/downloads')

    assertNoErrorRouting('/settings/account')

    expect(router.currentRoute.fullPath).toEqual('/settings/account')

    assertErrorRouting('/analytics')

    expect(router.currentRoute.fullPath).toEqual('/analytics')

    assertErrorRouting('/usage')

    expect(router.currentRoute.fullPath).toEqual('/usage')

    assertErrorRouting('/usage/2021-08')

    expect(router.currentRoute.fullPath).toEqual('/usage/2021-08')

    assertErrorRouting('/settings/game-config')

    expect(router.currentRoute.fullPath).toEqual('/settings/game-config')

    assertErrorRouting('/settings/users')

    expect(router.currentRoute.fullPath).toEqual('/settings/users')
  })
})

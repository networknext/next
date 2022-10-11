import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import NavBar from '@/components/NavBar.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { FlagPlugin } from '@/plugins/flags'
import { faBell } from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

describe('NavBar.vue', () => {
  const localVue = createLocalVue()

  const ICONS = [
    faBell
  ]

  library.add(...ICONS)

  // Mount FontAwesomeIcons
  localVue.component('font-awesome-icon', FontAwesomeIcon)

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)
  localVue.use(FlagPlugin, {
    flags: [],
    useAPI: false,
    apiService: {}
  })

  const defaultStore = {
    state: {
      allBuyers: [],
      userProfile: newDefaultProfile(),
      currentPage: 'map',
      filter: {
        companyCode: ''
      },
      killLoops: false,
      isAnonymousPlus: false,
      isAnonymous: false,
      isAdmin: false,
      isExplorer: false,
      hasBilling: false,
      isBuyer: false
    },
    getters: {
      allBuyers: (state: any) => state.allBuyers,
      currentPage: (state: any) => state.currentPage,
      currentFilter: (state: any) => state.filter,
      userProfile: (state: any) => state.userProfile,
      isAnonymousPlus: (state: any) => state.isAnonymousPlus,
      isAnonymous: (state: any) => state.isAnonymous,
      isBuyer: (state: any) => state.isBuyer,
      isExplorer: (state: any) => state.isExplorer,
      hasBilling: (state: any) => state.hasBilling,
      isAdmin: (state: any) => state.isAdmin,
      killLoops: (state: any) => state.killLoops
    },
    actions: {
      toggleKillLoops ({ commit }: any, killLoops: boolean) {
        commit('TOGGLE_KILL_LOOPS', killLoops)
      }
    },
    mutations: {
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      },
      UPDATE_CURRENT_PAGE (state: any, currentPage: string) {
        state.currentPage = currentPage
      },
      UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
        state.filter = newFilter
      },
      TOGGLE_KILL_LOOPS (state: any, killLoops: boolean) {
        state.killLoops = killLoops
      },
      UPDATE_IS_ANONYMOUSPLUS (state: any, isAnonymousPlus: boolean) {
        state.isAnonymousPlus = isAnonymousPlus
      },
      UPDATE_IS_ANONYMOUS (state: any, isAnonymous: boolean) {
        state.isAnonymous = isAnonymous
      },
      UPDATE_IS_BUYER (state: any, isBuyer: boolean) {
        state.isBuyer = isBuyer
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      },
      UPDATE_IS_OWNER (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      },
      UPDATE_IS_EXPLORER (state: any, isExplorer: boolean) {
        state.isExplorer = isExplorer
      },
      UPDATE_HAS_BILLING (state: any, hasBilling: boolean) {
        state.hasBilling = hasBilling
      },
      UPDATE_HAS_ANALYTICS (state: any, hasAnalytics: boolean) {
        state.hasAnalytics = hasAnalytics
      }
    }
  }

  const $route = {
    query: ''
  }

  const mocks = {
    $route
  }

  const stubs = [
    'router-link',
    'v-tour'
  ]

  it('mount default navbar', () => {
    process.env.VUE_APP_MODE = 'local'
    const versionSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchPortalVersion').mockImplementation(() => {
      return Promise.resolve({ sha: '123456789', commit_message: 'test commit message' })
    })

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(NavBar, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    localVue.nextTick()

    expect(versionSpy).toBeCalledTimes(1)

    versionSpy.mockReset()
    wrapper.destroy()
  })

  it('checks nav links - anonymous', () => {
    process.env.VUE_APP_MODE = 'local'
    const versionSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchPortalVersion').mockImplementation(() => {
      return Promise.resolve({ sha: '123456789', commit_message: 'test commit message' })
    })

    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_ANONYMOUS', true)

    const wrapper = shallowMount(NavBar, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    localVue.nextTick()

    expect(versionSpy).toBeCalledTimes(1)

    const links = wrapper.findAll('router-link-stub')
    expect(links.length).toBe(5)
    expect(links.at(0).text()).toBe('Map')
    expect(links.at(1).text()).toBe('Sessions')
    expect(links.at(2).text()).toBe('Session Tool')
    expect(links.at(3).text()).toBe('Log in')
    expect(links.at(4).text()).toBe('Get Access')

    store.commit('UPDATE_IS_ANONYMOUS', false)

    versionSpy.mockReset()
    wrapper.destroy()
  })

  it('checks nav links - anonymousPlus', () => {
    process.env.VUE_APP_MODE = 'local'
    const versionSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchPortalVersion').mockImplementation(() => {
      return Promise.resolve({ sha: '123456789', commit_message: 'test commit message' })
    })

    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ANONYMOUSPLUS', true)

    const profile = newDefaultProfile()
    profile.email = 'test@test.com'
    store.commit('UPDATE_USER_PROFILE', profile)

    const wrapper = shallowMount(NavBar, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    localVue.nextTick()

    expect(versionSpy).toBeCalledTimes(1)

    const links = wrapper.findAll('router-link-stub')
    expect(links.length).toBe(4)
    expect(links.at(0).text()).toBe('Map')
    expect(links.at(1).text()).toBe('Sessions')
    expect(links.at(2).text()).toBe('Session Tool')
    expect(links.at(3).text()).toBe('User Tool')

    const emailIndicator = wrapper.find('#email-indicator')
    expect(emailIndicator.exists()).toBeTruthy()
    expect(emailIndicator.text()).toBe(store.getters.userProfile.email)

    const logoutButton = wrapper.find('#logout-button')
    expect(logoutButton.exists()).toBeTruthy()

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    store.commit('UPDATE_IS_ANONYMOUSPLUS', false)

    versionSpy.mockReset()
    wrapper.destroy()
  })

  it('checks nav links - !anonymous - !anonymousPlus - !explorer', () => {
    process.env.VUE_APP_MODE = 'local'
    const versionSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchPortalVersion').mockImplementation(() => {
      return Promise.resolve({ sha: '123456789', commit_message: 'test commit message' })
    })

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(NavBar, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    localVue.nextTick()

    expect(versionSpy).toBeCalledTimes(1)

    const links = wrapper.findAll('router-link-stub')
    expect(links.length).toBe(6)
    expect(links.at(0).text()).toBe('Map')
    expect(links.at(1).text()).toBe('Sessions')
    expect(links.at(2).text()).toBe('Session Tool')
    expect(links.at(3).text()).toBe('User Tool')
    expect(links.at(4).text()).toBe('Downloads')
    expect(links.at(5).text()).toBe('Settings')

    const logoutButton = wrapper.find('#logout-button')
    expect(logoutButton.exists()).toBeTruthy()

    versionSpy.mockReset()
    wrapper.destroy()
  })

  it('checks nav links - !anonymous - !anonymousPlus - explorer', () => {
    process.env.VUE_APP_MODE = 'local'
    const versionSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchPortalVersion').mockImplementation(() => {
      return Promise.resolve({ sha: '123456789', commit_message: 'test commit message' })
    })

    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_EXPLORER', true)
    store.commit('UPDATE_HAS_BILLING', true)

    const wrapper = shallowMount(NavBar, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    localVue.nextTick()

    expect(versionSpy).toBeCalledTimes(1)

    const links = wrapper.findAll('router-link-stub')
    expect(links.length).toBe(7)
    expect(links.at(0).text()).toBe('Map')
    expect(links.at(1).text()).toBe('Sessions')
    expect(links.at(2).text()).toBe('Session Tool')
    expect(links.at(3).text()).toBe('User Tool')
    expect(links.at(4).text()).toBe('Downloads')
    expect(links.at(5).text()).toBe('Settings')
    expect(links.at(6).text()).toBe('Usage')

    const logoutButton = wrapper.find('#logout-button')
    expect(logoutButton.exists()).toBeTruthy()

    store.commit('UPDATE_HAS_BILLING', false)
    store.commit('UPDATE_IS_EXPLORER', false)

    versionSpy.mockReset()
    wrapper.destroy()
  })

  it('checks nav links - owner', () => {
    process.env.VUE_APP_MODE = 'local'
    const versionSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchPortalVersion').mockImplementation(() => {
      return Promise.resolve({ sha: '123456789', commit_message: 'test commit message' })
    })

    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_OWNER', true)

    const wrapper = shallowMount(NavBar, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    localVue.nextTick()

    expect(versionSpy).toBeCalledTimes(1)

    const links = wrapper.findAll('router-link-stub')
    expect(links.length).toBe(6)
    expect(links.at(0).text()).toBe('Map')
    expect(links.at(1).text()).toBe('Sessions')
    expect(links.at(2).text()).toBe('Session Tool')
    expect(links.at(3).text()).toBe('User Tool')
    expect(links.at(4).text()).toBe('Downloads')
    expect(links.at(5).text()).toBe('Settings')

    const notificationBell = wrapper.find('#notification-bell')
    expect(notificationBell.exists()).toBeTruthy()

    const logoutButton = wrapper.find('#logout-button')
    expect(logoutButton.exists()).toBeTruthy()

    store.commit('UPDATE_IS_OWNER', false)

    versionSpy.mockReset()
    wrapper.destroy()
  })
})

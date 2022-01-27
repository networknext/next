import { createLocalVue, mount, shallowMount } from '@vue/test-utils'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'
import Vuex from 'vuex'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AlertType } from '@/components/types/AlertTypes'
import { ErrorTypes } from '@/components/types/ErrorTypes'
import { FlagPlugin } from '@/plugins/flags'
import { FeatureEnum } from '@/components/types/FeatureTypes'

describe('SettingsWorkspace.vue', () => {
  const localVue = createLocalVue()

  // Setup plugins
  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)
  localVue.use(FlagPlugin, {
    flags: [],
    useAPI: false
  })

  // Init the store instance
  const defaultStore = {
    state: {
      currentPage: 'map',
      userProfile: newDefaultProfile(),
      isAnonymous: false,
      killLoops: false,
      isAnonymousPlus: false
    },
    getters: {
      currentPage: (state: any) => state.currentPage,
      userProfile: (state: any) => state.userProfile,
      isAdmin: (state: any) => state.userProfile.roles.indexOf('Admin') !== -1,
      isOwner: (state: any, getters: any) => state.userProfile.roles.indexOf('Owner') !== -1 || getters.isAdmin,
      registeredToCompany: (state: any) => (state.userProfile.companyCode !== ''),
      killLoops: (state: any) => state.killLoops
    },
    mutations: {
      TOGGLE_IS_ANONYMOUS (state: any, isAnonymous: boolean) {
        state.isAnonymous = isAnonymous
      },
      TOGGLE_IS_ANONYMOUSPLUS (state: any, isAnonymousPlus: boolean) {
        state.isAnonymousPlus = isAnonymousPlus
      },
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      },
      UPDATE_USER_ROLES (state: any, roles: Array<string>) {
        state.userProfile.roles = roles
      },
      UPDATE_USER_CUSTOMER_CODE (state: any, customerCode: string) {
        state.userProfile.companyCode = customerCode
      },
      TOGGLE_KILL_LOOPS (state: any, killLoops: boolean) {
        state.killLoops = killLoops
      },
      UPDATE_CURRENT_PAGE (state: any, currentPage: string) {
        state.currentPage = currentPage
      }
    }
  }

  const $route = {
    path: '/settings',
    params: {
      pathMatch: ''
    }
  }

  const mocks = {
    $route,
    $router: {
      push: (newRoute: any) => {
        $route.path = newRoute.path
      }
    }
  }

  const stubs = [
    'router-view',
    'router-link'
  ]

  it('mounts the settings workspace successfully', () => {
    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(SettingsWorkspace, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('checks layout', async () => {
    const store = new Vuex.Store(defaultStore)

    let wrapper = shallowMount(SettingsWorkspace, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    const cards = wrapper.findAll('.card')
    expect(cards.length).toBe(1)

    let tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(1)

    expect(tabs.at(0).text()).toBe('Account Settings')

    store.commit('UPDATE_USER_CUSTOMER_CODE', 'test')
    store.commit('UPDATE_USER_ROLES', ['Owner'])

    await localVue.nextTick()

    tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(3)

    expect(tabs.at(0).text()).toBe('Account Settings')
    expect(tabs.at(1).text()).toBe('Game Settings')
    expect(tabs.at(2).text()).toBe('Users')

    wrapper.destroy()

    localVue.prototype.$flagService.flags = [
      {
        name: FeatureEnum.FEATURE_ROUTE_SHADER,
        description: 'Route shader page for users to update their route shader',
        value: true
      }
    ]

    wrapper = shallowMount(SettingsWorkspace, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(4)

    expect(tabs.at(0).text()).toBe('Account Settings')
    expect(tabs.at(1).text()).toBe('Game Settings')
    expect(tabs.at(2).text()).toBe('Route Shader')
    expect(tabs.at(3).text()).toBe('Users')

    store.commit('UPDATE_USER_CUSTOMER_CODE', '')
    store.commit('UPDATE_USER_ROLES', [])

    wrapper.destroy()

    localVue.prototype.$flagService.flags = []
  })

  it('checks kill loops alert', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('TOGGLE_KILL_LOOPS', true)

    const wrapper = mount(SettingsWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe(ErrorTypes.SYSTEM_FAILURE)

    wrapper.destroy()

    store.commit('TOGGLE_KILL_LOOPS', false)

    wrapper.destroy()
  })
})

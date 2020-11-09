import { shallowMount, createLocalVue } from '@vue/test-utils'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'
import Vuex from 'vuex'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { FlagPlugin } from '@/plugins/flags'

describe('SettingsWorkspace.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)
  localVue.use(FlagPlugin, {
    flags: [
      {
        name: 'FEATURE_EXPLORE',
        description: 'Integrate Looker into the portal under a new navigation tab called "Explore"',
        value: 'false'
      },
      {
        name: 'FEATURE_INTERCOM',
        description: 'Integrate intercom',
        value: 'false'
      },
      {
        name: 'FEATURE_ROUTE_SHADER',
        description: 'Route shader page for users to update their route shader',
        value: 'false'
      },
      {
        name: 'FEATURE_IMPERSONATE',
        description: 'Feature to allow admins to impersonate a customer in a read only state',
        value: 'false'
      }
    ],
    useAPI: process.env.VUE_APP_USE_API_FLAGS
    // apiService: Vue.prototype.$apiService
  })

  const defaultStore = new Vuex.Store({
    state: {
      userProfile: {
        company: ''
      }
    },
    getters: {
      userProfile: (state: any) => state.userProfile,
      isAdmin: () => false,
      isOwner: () => false,
      registeredToCompany: () => false
    }
  })

  const $route = {
    path: '/session-tool',
    params: {
      pathMatch: ''
    }
  }

  const mocks = {
    $route,
    $router: {
      push: function (newRoute: any) {
        $route.path = newRoute.path
      }
    }
  }

  const stubs = [
    'router-view',
    'router-link'
  ]

  it('mounts the settings workspace successfully', () => {
    const store = defaultStore
    const wrapper = shallowMount(SettingsWorkspace, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })
})

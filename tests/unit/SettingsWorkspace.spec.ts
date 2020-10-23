import { shallowMount, createLocalVue } from '@vue/test-utils'
import SettingsWorkspace from '@/workspaces/SettingsWorkspace.vue'
import Vuex from 'vuex'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

describe('SettingsWorkspace.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

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

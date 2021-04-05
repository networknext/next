import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

describe('UserManagement.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = new Vuex.Store({
    state: {
      userProfile: {
        companyCode: '',
        companyName: '',
        domains: []
      }
    },
    getters: {
      userProfile: (state: any) => state.userProfile
    }
  })

  describe('MapWorkspace.vue', () => {
    it('mounts the users tab in the settings workspace successfully', () => {
      const store = defaultStore
      const wrapper = shallowMount(MapWorkspace, { localVue, store })
      expect(wrapper.exists()).toBe(true)
    })

    it('checks to see if everything is correct by default', () => {
      const store = defaultStore
      const wrapper = shallowMount(MapWorkspace, { localVue, store })
      expect(wrapper.exists()).toBe(true)
    })
  })
})

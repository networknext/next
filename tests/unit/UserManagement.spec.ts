import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import UserManagement from '@/components/UserManagement.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

describe('SessionCounts.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = new Vuex.Store({
    state: {
    },
    getters: {
    }
  })

  const spyAccounts = jest.spyOn(localVue.prototype.$apiService, 'fetchAllAccounts').mockImplementation(() => {
    return Promise.resolve({
      accounts: [

      ]
    })
  })

  const spyRoles = jest.spyOn(localVue.prototype.$apiService, 'fetchAllRoles').mockImplementation(() => {
    return Promise.resolve({
      roles: [

      ]
    })
  })

  describe('UserManagement.vue', () => {
    it('mounts the game config tab in the settings workspace successfully', () => {
      const store = defaultStore
      const wrapper = shallowMount(UserManagement, { localVue, store })
      expect(wrapper.exists()).toBe(true)
    })

    it('checks to see if everything is correct by default', () => {
      const store = defaultStore
      const wrapper = shallowMount(UserManagement, { localVue, store })
      expect(wrapper.exists()).toBe(true)
    })
  })
})

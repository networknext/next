import { shallowMount, createLocalVue } from '@vue/test-utils'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import Vuex from 'vuex'

describe('UserSessions.vue no sessions', () => {
  const localVue = createLocalVue()
  const store = new Vuex.Store({
    state: {
      filter: {
        buyerID: ''
      },
      showTable: false
    },
    getters: {
      currentFilter: (state: any) => state.filter
    },
    mutations: {
      TOGGLE_SESSION_TABLE (state: any, showTable: boolean) {
        state.showTable = showTable
      }
    }
  })

  localVue.use(JSONRPCPlugin)

  const spy = jest.spyOn(localVue.prototype.$apiService, 'fetchTopSessions').mockImplementationOnce(() => {
    return Promise.resolve({ sessions: [] })
  })

  it('mounts the user sessions table successfully', () => {
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store })
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })
})

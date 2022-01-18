import { shallowMount, createLocalVue } from '@vue/test-utils'
import SessionDetails from '@/components/SessionDetails.vue'
import Vuex from 'vuex'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile } from '@/components/types/AuthTypes'
import { VueConstructor } from 'vue/types/umd'

function fetchSessionDetailsMock (localVue: VueConstructor<any>, success: boolean, meta: any, slices: Array<any>, sessionID: string) {
  return jest.spyOn(localVue.prototype.$apiService, 'fetchSessionDetails').mockImplementation((args: any) => {
    expect(args.session_id).toBe(sessionID)
    return success ? Promise.resolve({
      meta: meta,
      slices: slices
    }) : Promise.reject(new Error('fetchSessionDetailsMock Error'))
  })
}

describe('SessionDetails.vue', () => {
  jest.useFakeTimers()
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  // Init the store instance
  const defaultStore = {
    state: {
      userProfile: newDefaultProfile(),
      filter: {
        companyCode: '',
        dateRange: DateFilterType.CURRENT_MONTH
      },
      killLoops: false,
      isAnonymous: true,
      isAnonymousPlus: false,
      viewport: null
    },
    getters: {
      userProfile: (state: any) => state.userProfile,
      killLoops: (state: any) => state.killLoops,
      isAnonymous: (state: any) => state.isAnonymous,
      isAnonymousPlus: (state: any) => state.isAnonymousPlus,
      currentFilter: (state: any) => state.filter,
      currentViewport: (state: any) => state.viewport
    },
    actions: {
      toggleKillLoops ({ commit }: any, killLoops: boolean) {
        commit('TOGGLE_KILL_LOOPS', killLoops)
      },
      updateCurrentViewport ({ commit }: any, viewport: any) {
        commit('UPDATE_CURRENT_VIEWPORT', viewport)
      }
    },
    mutations: {
      UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
        state.filter = newFilter
      },
      TOGGLE_KILL_LOOPS (state: any, killLoops: boolean) {
        state.killLoops = killLoops
      },
      UPDATE_IS_ANONYMOUS_PLUS (state: any, isAnonymousPlus: boolean) {
        state.isAnonymousPlus = isAnonymousPlus
      },
      UPDATE_IS_ANONYMOUS (state: any, isAnonymous: boolean) {
        state.isAnonymous = isAnonymous
      },
      UPDATE_CURRENT_VIEWPORT (state: any, viewport: any) {
        state.viewport = viewport
      }
    }
  }

  // Run bare minimum mount test
  it('mounts the map successfully', () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(SessionDetails, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    wrapper.destroy()
  })
})

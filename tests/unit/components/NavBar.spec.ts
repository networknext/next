import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import NavBar from '@/components/NavBar.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { FlagPlugin } from '@/plugins/flags'

describe('NavBar.vue', () => {
  const localVue = createLocalVue()

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
      UPDATE_IS_EXPLORER (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
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
    const versionSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchPortalVersion').mockImplementation(() => {
      return Promise.resolve({ sha: '123456789', commit_message: 'test commit message' })
    })

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(NavBar, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    localVue.nextTick()

    expect(versionSpy).toBeCalledTimes(1)

    wrapper.destroy()
  })
})

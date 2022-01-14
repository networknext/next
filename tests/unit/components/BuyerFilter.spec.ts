import Vuex from 'vuex'
import { createLocalVue, shallowMount } from '@vue/test-utils'
import BuyerFilter from '@/components/BuyerFilter.vue'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

describe('BuyerFilter.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = {
    state: {
      allBuyers: [],
      filter: {
        companyCode: '',
        dateRange: DateFilterType.CURRENT_MONTH
      },
      isAdmin: false,
      userProfile: newDefaultProfile()
    },
    getters: {
      allBuyers: (state: any) => state.allBuyers,
      currentFilter: (state: any) => state.filter,
      userProfile: (state: any) => state.userProfile,
      isAdmin: (state: any) => state.isAdmin
    },
    mutations: {
      UPDATE_ALL_BUYERS (state: any, allBuyers: Array<any>) {
        state.allBuyers = allBuyers
      },
      UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
        state.filter = newFilter
      },
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      }
    }
  }

  // Run bare minimum mount test
  it('mounts the component successfully', () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(BuyerFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    wrapper.destroy()
  })
})

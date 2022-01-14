import Vuex from 'vuex'
import { createLocalVue, shallowMount } from '@vue/test-utils'
import NotificationsModal from '@/components/NotificationsModal.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { Filter } from '@/components/types/FilterTypes'

describe('NotificationsModal.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = {
    state: {
      filter: {
        companyCode: ''
      },
      isAdmin: false,
      userProfile: newDefaultProfile()
    },
    getters: {
      currentFilter: (state: any) => state.filter,
      userProfile: (state: any) => state.userProfile,
      isAdmin: (state: any) => state.isAdmin
    },
    mutations: {
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

  it('mounts a notifications component successfully', () => {
    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

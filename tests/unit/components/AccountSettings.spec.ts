import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import AccountSettings from '@/components/AccountSettings.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'

describe('AccountSettings.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultProfile = newDefaultProfile()

  const store = new Vuex.Store({
    state: {
      allBuyers: [],
      userProfile: defaultProfile
    },
    getters: {
      allBuyers: (state: any) => state.allBuyers,
      isOwner: () => (state: any) => state.userProfile.roles.indexOf('Owner') !== -1,
      userProfile: (state: any) => state.userProfile
    },
    mutations: {
      UPDATE_ALL_BUYERS (state: any, allBuyers: Array<any>) {
        state.allBuyers = allBuyers
      },
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      }
    }
  })

  // Run bare minimum mount test
  it('mounts the component successfully', () => {
    const wrapper = shallowMount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

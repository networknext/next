import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import TermsOfServiceModal from '@/components/TermsOfServiceModal.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'

function buyerTOSSignedMock (vueInstance: VueConstructor<any>, success: boolean, firstName: string, lastName: string, email: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'buyerTOSSigned').mockImplementation((args: any) => {
    expect(args.first_name).toBe(firstName)
    expect(args.last_name).toBe(lastName)
    expect(args.email).toBe(email)

    return success ? Promise.resolve() : Promise.reject(new Error('buyerTOSSignedMock Mock Error'))
  })
}

describe('TermsOfServiceModal.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = {
    state: {
      isAdmin: false,
      isOwner: false,
      userProfile: newDefaultProfile()
    },
    getters: {
      userProfile: (state: any) => state.userProfile,
      isAdmin: (state: any) => state.isAdmin,
      isOwner: (state: any) => state.isOwner
    },
    mutations: {
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      },
      UPDATE_IS_OWNER (state: any, isOwner: boolean) {
        state.isOwner = isOwner
      }
    }
  }

  // Run bare minimum mount test
  it('mounts the component successfully', async () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(TermsOfServiceModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    wrapper.destroy()
  })

  it('checks for correct buttons - !deniable', async () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(TermsOfServiceModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const buttons = wrapper.findAll('button')

    expect(buttons.length).toBe(1)

    expect(buttons.at(0).text()).toBe('Accept')
    wrapper.destroy()
  })

  it('checks for correct buttons - deniable', async () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(TermsOfServiceModal, { localVue, store, propsData: { deniable: true } })
    expect(wrapper.exists()).toBeTruthy()

    const buttons = wrapper.findAll('button')

    expect(buttons.length).toBe(2)

    expect(buttons.at(0).text()).toBe('Deny')
    expect(buttons.at(1).text()).toBe('Accept')
    wrapper.destroy()
  })

  it('successful accept', async () => {
    const store = new Vuex.Store(defaultStore)

    const newUserProfile = newDefaultProfile()
    newUserProfile.firstName = 'test'
    newUserProfile.lastName = 'test'
    newUserProfile.email = 'test@test.com'

    store.commit('UPDATE_USER_PROFILE', newUserProfile)

    const spy = buyerTOSSignedMock(localVue, true, 'test', 'test', 'test@test.com')

    const wrapper = shallowMount(TermsOfServiceModal, { localVue, store, propsData: { deniable: true } })
    expect(wrapper.exists()).toBeTruthy()

    const buttons = wrapper.findAll('button')

    expect(buttons.length).toBe(2)

    expect(buttons.at(0).text()).toBe('Deny')
    expect(buttons.at(1).text()).toBe('Accept')

    await buttons.at(1).trigger('click')

    await localVue.nextTick()

    expect(spy).toBeCalled()

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())

    wrapper.destroy()
  })
})

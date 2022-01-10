import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import SessionCounts from '@/components/SessionCounts.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AlertType } from '@/components/types/AlertTypes'
import { ErrorTypes } from '@/components/types/ErrorTypes'

describe('UserManagement.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const store = new Vuex.Store({
    state: {
      allBuyers: [],
      userProfile: newDefaultProfile(),
      currentPage: 'map',
      filter: {
        companyCode: ''
      },
      killLoops: false,
      isAnonymousPlus: false,
      isAdmin: false,
      isBuyer: false
    },
    getters: {
      allBuyers: (state: any) => state.allBuyers,
      currentPage: (state: any) => state.currentPage,
      currentFilter: (state: any) => state.filter,
      userProfile: (state: any) => state.userProfile,
      isAnonymousPlus: (state: any) => state.isAnonymousPlus,
      isBuyer: (state: any) => state.isBuyer,
      isAdmin: (state: any) => state.isAdmin,
      killLoops: (state: any) => state.killLoops
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
      SET_SESSION_COUNT_ALERT_MESSAGE (state: any, sessionCountAlertMessage: string) {
        state.sessionCountAlertMessage = sessionCountAlertMessage
      },
      UPDATE_IS_ANONYMOUSPLUS (state: any, isAnonymousPlus: boolean) {
        state.isAnonymousPlus = isAnonymousPlus
      },
      UPDATE_IS_BUYER (state: any, isBuyer: boolean) {
        state.isBuyer = isBuyer
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      }
    }
  })

  jest.useFakeTimers()

  describe('SessionCounts.vue', () => {
    it('checks to see if everything is correct by default', () => {
      const spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
        return Promise.resolve({
          direct: 0,
          next: 0
        })
      })

      const wrapper = shallowMount(SessionCounts, { localVue, store })
      expect(wrapper.exists()).toBe(true)

      spyMapPoints.mockReset()

      wrapper.destroy()
    })

    it('checks count polling logic', async () => {
      const spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
        return Promise.resolve({
          direct: 0,
          next: 0
        })
      })

      const wrapper = shallowMount(SessionCounts, { localVue, store })
      expect(wrapper.exists()).toBeTruthy()

      expect(spyMapPoints).toBeCalled()
      expect(spyMapPoints).toBeCalledTimes(1)
      jest.advanceTimersByTime(1000)
      expect(spyMapPoints).toBeCalledTimes(2)
      jest.advanceTimersByTime(1000)
      expect(spyMapPoints).toBeCalledTimes(3)
      jest.advanceTimersByTime(1000)
      expect(spyMapPoints).toBeCalledTimes(4)

      spyMapPoints.mockReset()

      wrapper.destroy()
    })

    it('checks to see if loop interrupt works as expected', async () => {
      let spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation((args: any) => {
        expect(args.company_code).toBe('')

        return Promise.resolve({
          direct: 0,
          next: 0
        })
      })

      const wrapper = shallowMount(SessionCounts, { localVue, store })
      expect(wrapper.exists()).toBeTruthy()

      await wrapper.vm.$nextTick()

      spyMapPoints.mockReset()
      expect(spyMapPoints).not.toBeCalled()

      spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation((args: any) => {
        expect(args.company_code).toBe('test-company')

        return Promise.resolve({
          direct: 10,
          next: 100
        })
      })

      store.commit('UPDATE_CURRENT_FILTER', { companyCode: 'test-company' })

      await wrapper.vm.$nextTick()

      expect(spyMapPoints).toBeCalled()

      spyMapPoints.mockReset()
      wrapper.destroy()
    })
  })

  it('checks display update functionality', async () => {
    let spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
      return Promise.resolve({
        direct: 0,
        next: 0
      })
    })

    const wrapper = shallowMount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await wrapper.vm.$nextTick()

    expect(spyMapPoints).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    let headerElement = wrapper.find('h1')
    expect(headerElement.exists()).toBeTruthy()

    // Not the best way but doing this to avoid fighting with white space
    expect(headerElement.text()).toContain('Map')
    expect(headerElement.text()).toContain('0 Total Sessions')
    expect(headerElement.text()).toContain('0 on Network Next')

    spyMapPoints.mockReset()

    spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
      return Promise.resolve({
        direct: 10,
        next: 100
      })
    })

    jest.advanceTimersByTime(1000)
    expect(spyMapPoints).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    headerElement = wrapper.find('h1')
    expect(headerElement.exists()).toBeTruthy()

    // Not the best way but doing this to avoid fighting with white space
    expect(headerElement.text()).toContain('Map')
    expect(headerElement.text()).toContain('110 Total Sessions')
    expect(headerElement.text()).toContain('100 on Network Next')

    spyMapPoints.mockReset()

    wrapper.destroy()
  })

  // More of an integration test -> errors here will probably be thrown by BuyerFilter component
  it('checks buyer filter', async () => {
    const spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
      return Promise.resolve({
        direct: 0,
        next: 0
      })
    })

    const wrapper = mount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await wrapper.vm.$nextTick()

    expect(spyMapPoints).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    let buyerFilter = wrapper.find('#buyer-filter')
    expect(buyerFilter.exists()).toBeFalsy()

    const defaultProfile = newDefaultProfile()
    defaultProfile.pubKey = 'test pub key'
    store.commit('UPDATE_IS_BUYER', true)

    await wrapper.vm.$nextTick()

    buyerFilter = wrapper.find('#buyer-filter')
    expect(buyerFilter.exists()).toBeTruthy()

    store.commit('UPDATE_IS_BUYER', false)
    await wrapper.vm.$nextTick()

    buyerFilter = wrapper.find('#buyer-filter')
    expect(buyerFilter.exists()).toBeFalsy()

    store.commit('UPDATE_IS_ADMIN', true)
    await wrapper.vm.$nextTick()

    buyerFilter = wrapper.find('#buyer-filter')
    expect(buyerFilter.exists()).toBeTruthy()

    expect(store.getters.isBuyer).toBeFalsy()
    expect(store.getters.isAdmin).toBeTruthy()

    spyMapPoints.mockReset()

    wrapper.destroy()
  })

  it('checks verification alert', async () => {
    const spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
      return Promise.resolve({
        direct: 0,
        next: 0
      })
    })

    store.commit('UPDATE_IS_ANONYMOUSPLUS', true)

    const profile = newDefaultProfile()
    profile.email = 'test@test.com'

    store.commit('UPDATE_USER_PROFILE', profile)
    store.commit('UPDATE_IS_ANONYMOUSPLUS', true)

    const wrapper = mount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await wrapper.vm.$nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.INFO)).toBeTruthy()
    expect(alert.text()).toContain('Please check your email to verify your email address: test@test.com')
    expect(alert.text()).toContain('Resend email')

    spyMapPoints.mockReset()

    wrapper.destroy()
  })

  it('checks kill loops alert', async () => {
    const spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
      return Promise.resolve({
        direct: 0,
        next: 0
      })
    })

    let wrapper = mount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    let alert = wrapper.find('.alert')
    expect(alert.exists()).toBeFalsy()

    store.commit('TOGGLE_KILL_LOOPS', true)

    await wrapper.vm.$nextTick()

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe(ErrorTypes.SYSTEM_FAILURE)

    wrapper.destroy()

    wrapper = mount(SessionCounts, { localVue, store })

    await wrapper.vm.$nextTick()

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe(ErrorTypes.SYSTEM_FAILURE)

    spyMapPoints.mockReset()

    wrapper.destroy()
  })

  it('checks failed map lookup alert', async () => {
    const spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
      return Promise.resolve({
        direct: 0,
        next: 0
      })
    })

    const wrapper = mount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    wrapper.vm.$root.$emit('failedMapPointLookup')

    await wrapper.vm.$nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.WARNING)).toBeTruthy()
    expect(alert.text()).toBe(ErrorTypes.FAILED_MAP_POINT_LOOKUP)

    spyMapPoints.mockReset()

    wrapper.destroy()
  })
})

import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import SessionCounts from '@/components/SessionCounts.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AlertType } from '@/components/types/AlertTypes'
import { ErrorTypes } from '@/components/types/ErrorTypes'
import { VueConstructor } from 'vue/types/umd'
import { MAX_RETRIES } from '@/components/types/Constants'

function totalSessionCountsMock (vueInstance: VueConstructor<any>, success: boolean, direct: number, next: number, customerCode: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation((args: any) => {
    expect(args.company_code).toBe(customerCode)

    return success ? Promise.resolve(
      {
        direct: direct,
        next: next
      }
    ) : Promise.reject()
  })
}

describe('SessionCounts.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

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
  }

  jest.useFakeTimers()

  it('checks to see if everything is correct by default', () => {
    const spy = totalSessionCountsMock(localVue, true, 0, 0, '')
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBe(true)

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks count polling logic', async () => {
    const spy = totalSessionCountsMock(localVue, true, 0, 0, '')
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(spy).toBeCalled()
    expect(spy).toBeCalledTimes(1)
    jest.advanceTimersByTime(1000)
    expect(spy).toBeCalledTimes(2)
    jest.advanceTimersByTime(1000)
    expect(spy).toBeCalledTimes(3)
    jest.advanceTimersByTime(1000)
    expect(spy).toBeCalledTimes(4)

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks to see if loop interrupt works as expected', async () => {
    let spy = totalSessionCountsMock(localVue, true, 0, 0, '')
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await wrapper.vm.$nextTick()

    spy.mockReset()
    expect(spy).not.toBeCalled()

    spy = totalSessionCountsMock(localVue, true, 10, 100, 'test-company')

    store.commit('UPDATE_CURRENT_FILTER', { companyCode: 'test-company' })

    await wrapper.vm.$nextTick()

    expect(spy).toBeCalled()

    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '' })

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks display update functionality', async () => {
    let spy = totalSessionCountsMock(localVue, true, 0, 0, '')
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await wrapper.vm.$nextTick()

    expect(spy).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    let headerElement = wrapper.find('h1')
    expect(headerElement.exists()).toBeTruthy()

    // Not the best way but doing this to avoid fighting with white space
    expect(headerElement.text()).toContain('Map')
    expect(headerElement.text()).toContain('0 Total Sessions')
    expect(headerElement.text()).toContain('0 on Network Next')

    spy.mockReset()

    spy = totalSessionCountsMock(localVue, true, 10, 100, '')

    jest.advanceTimersByTime(1000)
    expect(spy).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    headerElement = wrapper.find('h1')
    expect(headerElement.exists()).toBeTruthy()

    // Not the best way but doing this to avoid fighting with white space
    expect(headerElement.text()).toContain('Map')
    expect(headerElement.text()).toContain('110 Total Sessions')
    expect(headerElement.text()).toContain('100 on Network Next')

    spy.mockReset()

    wrapper.destroy()
  })

  // More of an integration test -> errors here will probably be thrown by BuyerFilter component
  it('checks buyer filter', async () => {
    const spy = totalSessionCountsMock(localVue, true, 0, 0, '')
    const store = new Vuex.Store(defaultStore)

    const wrapper = mount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await wrapper.vm.$nextTick()

    expect(spy).toBeCalledTimes(1)

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

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks verification alert', async () => {
    const spy = totalSessionCountsMock(localVue, true, 0, 0, '')
    const store = new Vuex.Store(defaultStore)

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

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks kill loops alert', async () => {
    const spy = totalSessionCountsMock(localVue, true, 0, 0, '')
    const store = new Vuex.Store(defaultStore)

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

    store.commit('TOGGLE_KILL_LOOPS', false)

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks failed map lookup alert', async () => {
    const spy = totalSessionCountsMock(localVue, true, 0, 0, '')
    const store = new Vuex.Store(defaultStore)

    const wrapper = mount(SessionCounts, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    wrapper.vm.$root.$emit('failedMapPointLookup')

    await wrapper.vm.$nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.WARNING)).toBeTruthy()
    expect(alert.text()).toBe(ErrorTypes.FAILED_MAP_POINT_LOOKUP)

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks failed api call', async () => {
    const store = new Vuex.Store(defaultStore)

    const spy = totalSessionCountsMock(localVue, false, 0, 0, '')

    const wrapper = mount(SessionCounts, { localVue, store })

    expect(store.getters.killLoops).toBeFalsy()

    await wrapper.vm.$nextTick()

    expect(spy).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    let retryCount = wrapper.vm.$data.retryCount
    expect(retryCount).toBe(1)

    for (let i = 2; i <= MAX_RETRIES; i++) {
      jest.advanceTimersByTime(3000 * retryCount)

      await wrapper.vm.$nextTick()

      expect(spy).toBeCalledTimes(i)

      await wrapper.vm.$nextTick()

      retryCount = wrapper.vm.$data.retryCount
      expect(retryCount).toBe(i)
    }

    expect(store.getters.killLoops).toBeTruthy()

    wrapper.destroy()
  })
})

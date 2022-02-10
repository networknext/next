import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import Analytics from '@/components/Analytics.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'

function fetchAnalyticsDashboardsMock (vueInstance: VueConstructor<any>, success: boolean, urls: Array<string>, customerCode: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchAnalyticsDashboards').mockImplementation((args: any) => {
    expect(args.customer_code).toBe(customerCode)
    return success ? Promise.resolve({ urls: urls }) : Promise.reject(new Error('fetchAnalyticsDashboardsMock Mock Error'))
  })
}

describe('Analytics.vue', () => {
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

  // Run bare minimum mount test
  it('mounts the component successfully', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, [], '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a single dashboard category and one dashboard', async () => {})
  it('mounts a single dashboard category and multiple dashboards', async () => {})
  it('mounts a multiple dashboard category and one dashboard', async () => {})
  it('mounts a multiple dashboard category and multiple dashboards', async () => {})

  /*

  it('mounts a single analytics dashboard', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, ['https://127.0.0.1'], '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('analyticsDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts multiple analytics dashboards', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, [
      'https://127.0.0.1', 'https://127.0.0.2'
    ], '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(2)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('analyticsDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    expect(lookEmbeds.at(1).attributes('dashid')).toBe('analyticsDash')
    expect(lookEmbeds.at(1).attributes('dashurl')).toBe('https://127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('checks filter change update - !admin', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, ['https://127.0.0.1'], '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    let lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('analyticsDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    const newFilter = { companyCode: 'test' }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('analyticsDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '' })

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('checks filter change update - admin', async () => {
    let analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, ['https://127.0.0.1'], '')

    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    let lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('analyticsDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, ['https://127.0.0.2'], 'test')

    const newFilter = { companyCode: 'test' }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(2)

    await localVue.nextTick()

    lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('analyticsDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.2')

    wrapper.destroy()

    store.commit('UPDATE_IS_ADMIN', false)
    store.commit('UPDATE_CURRENT_FILTER', { company_code: '' })
    analyticDashSpy.mockReset()
  })

  it('checks failure alert', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, false, [], '')

    const wrapper = mount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe('Failed to fetch analytics dashboards. Please refresh the page')

    analyticDashSpy.mockReset()

    wrapper.destroy()
  })
  */
})

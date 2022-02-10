import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import Usage from '@/components/Usage.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AlertType } from '@/components/types/AlertTypes'

function fetchUsageSummaryMock (vueInstance: VueConstructor<any>, success: boolean, url: string, customerCode: string, dateString: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchUsageSummary').mockImplementation((args: any) => {
    expect(args.customer_code).toBe(customerCode)
    expect(args.date_string).toBe(dateString)
    return success ? Promise.resolve({ url: url }) : Promise.reject(new Error('fetchUsageSummaryMock Mock Error'))
  })
}

describe('Usage.vue', () => {
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

  const $route = {
    path: '/explore/usage',
    params: {
      pathMatch: ''
    }
  }

  const mocks = {
    $route,
    $router: {
      push: (newRoute: any) => {
        $route.path = newRoute.path
      }
    }
  }

  // Run bare minimum mount test
  it('mounts the component successfully', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchUsageSummaryMock(localVue, true, '', '', '')

    const wrapper = shallowMount(Usage, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a single Usage dashboard', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchUsageSummaryMock(localVue, true, 'https://127.0.0.1', '', '')

    const wrapper = shallowMount(Usage, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('usageDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('checks filter change update - !admin', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchUsageSummaryMock(localVue, true, 'https://127.0.0.1', '', '')

    const wrapper = shallowMount(Usage, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    let lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('usageDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    const newFilter = { companyCode: 'test' }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('usageDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '' })

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('checks filter change update - admin', async () => {
    let analyticDashSpy = fetchUsageSummaryMock(localVue, true, 'https://127.0.0.1', '', '')

    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)

    const wrapper = shallowMount(Usage, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    let lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('usageDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    analyticDashSpy = fetchUsageSummaryMock(localVue, true, 'https://127.0.0.2', 'test', '')

    const newFilter = { companyCode: 'test' }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(2)

    await localVue.nextTick()

    lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('usageDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.2')

    wrapper.destroy()

    store.commit('UPDATE_IS_ADMIN', false)
    store.commit('UPDATE_CURRENT_FILTER', { company_code: '' })
    analyticDashSpy.mockReset()
  })

  it('checks failure alert', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchUsageSummaryMock(localVue, false, '', '', '')

    const wrapper = mount(Usage, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe('Failed to fetch usage dashboard. Please refresh the page')

    analyticDashSpy.mockReset()

    wrapper.destroy()
  })

  // TODO: Add in checks for payment instructions
})

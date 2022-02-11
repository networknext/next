import Vuex from 'vuex'
import { createLocalVue, mount, shallowMount } from '@vue/test-utils'
import Discovery from '@/components/Discovery.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { Filter } from '@/components/types/FilterTypes'
import { VueConstructor } from 'vue/types/umd'

function fetchDiscoveryDashboardsMock (vueInstance: VueConstructor<any>, success: boolean, urls: Array<string>, customerCode: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchDiscoveryDashboards').mockImplementation((args: any) => {
    expect(args.customer_code).toBe(customerCode)
    return success ? Promise.resolve({ urls: urls }) : Promise.reject(new Error('fetchDiscoveryDashboardsMock Mock Error'))
  })
}

describe('Discovery.vue', () => {
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
    const discoveryDashSpy = fetchDiscoveryDashboardsMock(localVue, true, [], '')

    const wrapper = shallowMount(Discovery, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(discoveryDashSpy).toBeCalledTimes(1)

    const savesStub = wrapper.find('saves-stub')
    expect(savesStub.exists()).toBeTruthy()

    wrapper.destroy()

    discoveryDashSpy.mockReset()
  })

  it('mounts a single analytics dashboard', async () => {
    const store = new Vuex.Store(defaultStore)
    const discoveryDashSpy = fetchDiscoveryDashboardsMock(localVue, true, ['https://127.0.0.1'], '')

    const wrapper = shallowMount(Discovery, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(discoveryDashSpy).toBeCalledTimes(1)

    const lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('discoveryDashboard')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    const cardTitle = wrapper.find('.card-title')
    expect(cardTitle.exists()).toBeTruthy()
    expect(cardTitle.text()).toBe('Discovery Dashboards')

    const cardDetails = wrapper.find('.card-text')
    expect(cardDetails.exists()).toBeTruthy()
    expect(cardDetails.text()).toBe('Interesting one off dashboards that are created by the Network Next datascience team')

    wrapper.destroy()

    discoveryDashSpy.mockReset()
  })

  it('mounts multiple analytics dashboards', async () => {
    const store = new Vuex.Store(defaultStore)
    const discoveryDashSpy = fetchDiscoveryDashboardsMock(localVue, true, [
      'https://127.0.0.1', 'https://127.0.0.2'
    ], '')

    const wrapper = shallowMount(Discovery, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(discoveryDashSpy).toBeCalledTimes(1)

    const lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(2)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('discoveryDashboard')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    expect(lookEmbeds.at(1).attributes('dashid')).toBe('discoveryDashboard')
    expect(lookEmbeds.at(1).attributes('dashurl')).toBe('https://127.0.0.2')

    wrapper.destroy()

    discoveryDashSpy.mockReset()
  })

  it('checks filter change update - !admin', async () => {
    const store = new Vuex.Store(defaultStore)
    const discoveryDashSpy = fetchDiscoveryDashboardsMock(localVue, true, ['https://127.0.0.1'], '')

    const wrapper = shallowMount(Discovery, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(discoveryDashSpy).toBeCalledTimes(1)

    let lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('discoveryDashboard')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    const newFilter = { companyCode: 'test' }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()

    expect(discoveryDashSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('discoveryDashboard')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '' })

    wrapper.destroy()

    discoveryDashSpy.mockReset()
  })

  it('checks filter change update - admin', async () => {
    let discoveryDashSpy = fetchDiscoveryDashboardsMock(localVue, true, ['https://127.0.0.1'], '')

    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)

    const wrapper = shallowMount(Discovery, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(discoveryDashSpy).toBeCalledTimes(1)

    let lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('discoveryDashboard')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    discoveryDashSpy = fetchDiscoveryDashboardsMock(localVue, true, ['https://127.0.0.2'], 'test')

    const newFilter = { companyCode: 'test' }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()

    expect(discoveryDashSpy).toBeCalledTimes(2)

    await localVue.nextTick()

    lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('discoveryDashboard')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.2')

    wrapper.destroy()

    store.commit('UPDATE_IS_ADMIN', false)
    store.commit('UPDATE_CURRENT_FILTER', { company_code: '' })
    discoveryDashSpy.mockReset()
  })
})

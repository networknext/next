import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import Analytics from '@/components/Analytics.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'

export interface AnalyticsDashboards {
  [tab: string]: Array<string>;
}

function fetchAnalyticsDashboardsMock (vueInstance: VueConstructor<any>, success: boolean, dashboards: AnalyticsDashboards, customerCode: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchAnalyticsDashboards').mockImplementation((args: any) => {
    expect(args.customer_code).toBe(customerCode)

    return success ? Promise.resolve({ dashboards: dashboards }) : Promise.reject(new Error('fetchAnalyticsDashboardsMock Mock Error'))
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
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {}, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a single dashboard category and one dashboard', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      'General': [
        '127.0.0.1'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(1)
    expect(tabs.at(0).text()).toBe('General')

    const selectedTab = wrapper.findAll('.blue-accent')
    expect(selectedTab.length).toBe(1)
    expect(selectedTab.at(0).text()).toBe('General')

    const dashboards = wrapper.findAll('lookerembed-stub')
    expect(dashboards.length).toBe(1)
    expect(dashboards.at(0).attributes('dashurl')).toBe('127.0.0.1')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a single dashboard category and multiple dashboards', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      'General': [
        '127.0.0.1',
        '127.0.0.2'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(1)
    expect(tabs.at(0).text()).toBe('General')

    const selectedTab = wrapper.findAll('.blue-accent')
    expect(selectedTab.length).toBe(1)
    expect(selectedTab.at(0).text()).toBe('General')

    const dashboards = wrapper.findAll('lookerembed-stub')
    expect(dashboards.length).toBe(2)
    expect(dashboards.at(0).attributes('dashurl')).toBe('127.0.0.1')
    expect(dashboards.at(1).attributes('dashurl')).toBe('127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a multiple dashboard categories and one dashboard', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      'General': [
        '127.0.0.1'
      ],
      'Platform': [
        '127.0.0.2'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(2)
    expect(tabs.at(0).text()).toBe('General')
    expect(tabs.at(1).text()).toBe('Platform')

    const selectedTab = wrapper.findAll('.blue-accent')
    expect(selectedTab.length).toBe(1)
    expect(selectedTab.at(0).text()).toBe('General')

    const dashboards = wrapper.findAll('lookerembed-stub')
    expect(dashboards.length).toBe(1)
    expect(dashboards.at(0).attributes('dashurl')).toBe('127.0.0.1')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a multiple dashboard categories and one dashboard - General not first', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      'Platform': [
        '127.0.0.1'
      ],
      'General': [
        '127.0.0.2'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(2)
    expect(tabs.at(0).text()).toBe('General')
    expect(tabs.at(1).text()).toBe('Platform')

    const selectedTab = wrapper.findAll('.blue-accent')
    expect(selectedTab.length).toBe(1)
    expect(selectedTab.at(0).text()).toBe('General')

    const dashboards = wrapper.findAll('lookerembed-stub')
    expect(dashboards.length).toBe(1)
    expect(dashboards.at(0).attributes('dashurl')).toBe('127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('check tab switching', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      'General': [
        '127.0.0.1'
      ],
      'Platform': [
        '127.0.0.2'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(2)
    expect(tabs.at(0).text()).toBe('General')
    expect(tabs.at(1).text()).toBe('Platform')

    let selectedTab = wrapper.findAll('.blue-accent')
    expect(selectedTab.length).toBe(1)
    expect(selectedTab.at(0).text()).toBe('General')

    let dashboards = wrapper.findAll('lookerembed-stub')
    expect(dashboards.length).toBe(1)
    expect(dashboards.at(0).attributes('dashurl')).toBe('127.0.0.1')

    await tabs.at(1).trigger('click')

    await localVue.nextTick()

    selectedTab = wrapper.findAll('.blue-accent')
    expect(selectedTab.length).toBe(1)
    expect(selectedTab.at(0).text()).toBe('Platform')

    dashboards = wrapper.findAll('lookerembed-stub')
    expect(dashboards.length).toBe(1)
    expect(dashboards.at(0).attributes('dashurl')).toBe('127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })


  it('mounts a multiple dashboard categories and multiple dashboards', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      'General': [
        '127.0.0.1',
        '127.0.0.2'
      ],
      'Platform': [
        '127.0.0.3',
        '127.0.0.4'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const tabs = wrapper.findAll('li')
    expect(tabs.length).toBe(2)
    expect(tabs.at(0).text()).toBe('General')
    expect(tabs.at(1).text()).toBe('Platform')

    const selectedTab = wrapper.findAll('.blue-accent')
    expect(selectedTab.length).toBe(1)
    expect(selectedTab.at(0).text()).toBe('General')

    const dashboards = wrapper.findAll('lookerembed-stub')
    expect(dashboards.length).toBe(2)
    expect(dashboards.at(0).attributes('dashurl')).toBe('127.0.0.1')
    expect(dashboards.at(1).attributes('dashurl')).toBe('127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })
})

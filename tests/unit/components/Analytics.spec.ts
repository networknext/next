import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import Analytics from '@/components/Analytics.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'

interface AnalyticsDashboards {
  [tab: string]: Array<string>;
}

interface SubCategoryMap {
  [subTab: string]: any
}

function fetchAnalyticsDashboardsMock (vueInstance: VueConstructor<any>, success: boolean, dashboards: AnalyticsDashboards, tabs: Array<string>, subTabs: SubCategoryMap, customerCode: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchAnalyticsDashboards').mockImplementation((args: any) => {
    expect(args.customer_code).toBe(customerCode)

    return success ? Promise.resolve({ dashboards: dashboards, tabs: tabs, sub_tabs: subTabs }) : Promise.reject(new Error('fetchAnalyticsDashboardsMock Mock Error'))
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
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {}, [], {}, '')

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
      General: [
        '127.0.0.1'
      ]
    }, ['General'],
    {}, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(1)
    expect(mainTabs.at(0).text()).toBe('General')

    const activeTab = wrapper.find('.active')
    expect(activeTab.text()).toBe('General')

    const lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a single dashboard category and multiple dashboards', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      General: [
        '127.0.0.1',
        '127.0.0.2'
      ]
    }, ['General'],
    {}, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(1)
    expect(mainTabs.at(0).text()).toBe('General')

    const activeTab = wrapper.find('.active')
    expect(activeTab.text()).toBe('General')

    const lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(2)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')
    expect(lookerStubs.at(1).attributes('dashurl')).toBe('127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts multiple dashboard categories and one dashboard each', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      General: [
        '127.0.0.1'
      ],
      Platform: [
        '127.0.0.2'
      ]
    }, ['General', 'Platform'], {}, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('General')
    expect(mainTabs.at(1).text()).toBe('Platform')

    const activeTab = wrapper.find('.active')
    expect(activeTab.text()).toBe('General')

    const lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts multiple dashboard categories and one dashboard each - different order', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      General: [
        '127.0.0.1'
      ],
      Platform: [
        '127.0.0.2'
      ]
    }, ['Platform', 'General'], {}, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('Platform')
    expect(mainTabs.at(1).text()).toBe('General')

    const activeTab = wrapper.find('.active')
    expect(activeTab.text()).toBe('Platform')

    const lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a multiple dashboard categories and multiple dashboards each', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      Platform: [
        '127.0.0.1',
        '127.0.0.2',
        '127.0.0.3'
      ],
      General: [
        '127.0.0.4',
        '127.0.0.5',
        '127.0.0.6'
      ]
    }, ['Platform', 'General'], {}, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('Platform')
    expect(mainTabs.at(1).text()).toBe('General')

    const activeTab = wrapper.find('.active')
    expect(activeTab.text()).toBe('Platform')

    const lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(3)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')
    expect(lookerStubs.at(1).attributes('dashurl')).toBe('127.0.0.2')
    expect(lookerStubs.at(2).attributes('dashurl')).toBe('127.0.0.3')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('check tab switching', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      Platform: [
        '127.0.0.1',
        '127.0.0.2',
        '127.0.0.3'
      ],
      General: [
        '127.0.0.4',
        '127.0.0.5',
        '127.0.0.6'
      ]
    }, ['Platform', 'General'], {}, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    let mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('Platform')
    expect(mainTabs.at(1).text()).toBe('General')

    let activeTab = wrapper.find('.active')
    expect(activeTab.text()).toBe('Platform')

    let lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(3)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')
    expect(lookerStubs.at(1).attributes('dashurl')).toBe('127.0.0.2')
    expect(lookerStubs.at(2).attributes('dashurl')).toBe('127.0.0.3')

    await mainTabs.at(1).trigger('click')

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(2)

    mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('Platform')
    expect(mainTabs.at(1).text()).toBe('General')

    activeTab = wrapper.find('.active')
    expect(activeTab.text()).toBe('General')

    lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(3)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.4')
    expect(lookerStubs.at(1).attributes('dashurl')).toBe('127.0.0.5')
    expect(lookerStubs.at(2).attributes('dashurl')).toBe('127.0.0.6')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('mounts a single category with two sub categories and one dashboard per category', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      'Retention/Latency': [
        '127.0.0.1',
      ],
      'Retention/Region': [
        '127.0.0.2'
      ]
    }, ['Retention'],
    {
      Retention: [
        'Latency',
        'Region'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    const mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(1)
    expect(mainTabs.at(0).text()).toBe('Retention')

    const activeMainTab = wrapper.find('.active')
    expect(activeMainTab.exists()).toBeTruthy()
    expect(activeMainTab.text()).toBe('Retention')

    const subTabs = wrapper.findAll('.sub-li')
    expect(subTabs.length).toBe(2)
    expect(subTabs.at(0).text()).toBe('Latency')
    expect(subTabs.at(1).text()).toBe('Region')

    const activeSubTab = wrapper.find('.blue-accent')
    expect(activeSubTab.exists()).toBeTruthy()
    expect(activeSubTab.text()).toBe('Latency')

    const lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('checks switching between sub tabs', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      'Retention/Latency': [
        '127.0.0.1',
      ],
      'Retention/Region': [
        '127.0.0.2'
      ]
    }, ['Retention'],
    {
      Retention: [
        'Latency',
        'Region'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    let mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(1)
    expect(mainTabs.at(0).text()).toBe('Retention')

    let activeMainTab = wrapper.find('.active')
    expect(activeMainTab.exists()).toBeTruthy()
    expect(activeMainTab.text()).toBe('Retention')

    let subTabs = wrapper.findAll('.sub-li')
    expect(subTabs.length).toBe(2)
    expect(subTabs.at(0).text()).toBe('Latency')
    expect(subTabs.at(1).text()).toBe('Region')

    let activeSubTab = wrapper.find('.blue-accent')
    expect(activeSubTab.exists()).toBeTruthy()
    expect(activeSubTab.text()).toBe('Latency')

    let lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')

    await subTabs.at(1).trigger('click')

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(2)

    mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(1)
    expect(mainTabs.at(0).text()).toBe('Retention')

    activeMainTab = wrapper.find('.active')
    expect(activeMainTab.exists()).toBeTruthy()
    expect(activeMainTab.text()).toBe('Retention')

    subTabs = wrapper.findAll('.sub-li')
    expect(subTabs.length).toBe(2)
    expect(subTabs.at(0).text()).toBe('Latency')
    expect(subTabs.at(1).text()).toBe('Region')

    activeSubTab = wrapper.find('.blue-accent')
    expect(activeSubTab.exists()).toBeTruthy()
    expect(activeSubTab.text()).toBe('Region')

    lookerStubs = wrapper.findAll('lookerembed-stub')

    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })

  it('checks switching between sub tabs with other main tabs present', async () => {
    const store = new Vuex.Store(defaultStore)
    const analyticDashSpy = fetchAnalyticsDashboardsMock(localVue, true, {
      General: [
        '127.0.0.3'
      ],
      'Retention/Latency': [
        '127.0.0.1',
      ],
      'Retention/Region': [
        '127.0.0.2'
      ]
    }, ['Retention', 'General'],
    {
      Retention: [
        'Latency',
        'Region'
      ]
    }, '')

    const wrapper = shallowMount(Analytics, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(1)

    let mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('Retention')
    expect(mainTabs.at(1).text()).toBe('General')

    let activeMainTab = wrapper.find('.active')
    expect(activeMainTab.exists()).toBeTruthy()
    expect(activeMainTab.text()).toBe('Retention')

    let subTabs = wrapper.findAll('.sub-li')
    expect(subTabs.length).toBe(2)
    expect(subTabs.at(0).text()).toBe('Latency')
    expect(subTabs.at(1).text()).toBe('Region')

    let activeSubTab = wrapper.find('.blue-accent')
    expect(activeSubTab.exists()).toBeTruthy()
    expect(activeSubTab.text()).toBe('Latency')

    let lookerStubs = wrapper.findAll('lookerembed-stub')
    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')

    await mainTabs.at(1).trigger('click')

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(2)

    mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('Retention')
    expect(mainTabs.at(1).text()).toBe('General')

    activeMainTab = wrapper.find('.active')
    expect(activeMainTab.exists()).toBeTruthy()
    expect(activeMainTab.text()).toBe('General')

    subTabs = wrapper.findAll('.sub-li')
    expect(subTabs.length).toBe(0)

    activeSubTab = wrapper.find('.blue-accent')
    expect(activeSubTab.exists()).toBeFalsy()

    lookerStubs = wrapper.findAll('lookerembed-stub')
    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.3')

    await mainTabs.at(0).trigger('click')

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(3)

    mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('Retention')
    expect(mainTabs.at(1).text()).toBe('General')

    activeMainTab = wrapper.find('.active')
    expect(activeMainTab.exists()).toBeTruthy()
    expect(activeMainTab.text()).toBe('Retention')

    subTabs = wrapper.findAll('.sub-li')
    expect(subTabs.length).toBe(2)
    expect(subTabs.at(0).text()).toBe('Latency')
    expect(subTabs.at(1).text()).toBe('Region')

    activeSubTab = wrapper.find('.blue-accent')
    expect(activeSubTab.exists()).toBeTruthy()
    expect(activeSubTab.text()).toBe('Latency')

    lookerStubs = wrapper.findAll('lookerembed-stub')
    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.1')

    await subTabs.at(1).trigger('click')

    await localVue.nextTick()

    expect(analyticDashSpy).toBeCalledTimes(4)

    mainTabs = wrapper.findAll('.nav-link')
    expect(mainTabs.length).toBe(2)
    expect(mainTabs.at(0).text()).toBe('Retention')
    expect(mainTabs.at(1).text()).toBe('General')

    activeMainTab = wrapper.find('.active')
    expect(activeMainTab.exists()).toBeTruthy()
    expect(activeMainTab.text()).toBe('Retention')

    subTabs = wrapper.findAll('.sub-li')
    expect(subTabs.length).toBe(2)
    expect(subTabs.at(0).text()).toBe('Latency')
    expect(subTabs.at(1).text()).toBe('Region')

    activeSubTab = wrapper.find('.blue-accent')
    expect(activeSubTab.exists()).toBeTruthy()
    expect(activeSubTab.text()).toBe('Region')

    lookerStubs = wrapper.findAll('lookerembed-stub')
    expect(lookerStubs.length).toBe(1)
    expect(lookerStubs.at(0).attributes('dashurl')).toBe('127.0.0.2')

    wrapper.destroy()

    analyticDashSpy.mockReset()
  })
})

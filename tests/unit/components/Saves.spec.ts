import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import Saves from '@/components/Saves.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'

function fetchSavesDashboardMock (vueInstance: VueConstructor<any>, success: boolean, url: string, customerCode: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchSavesDashboard').mockImplementation((args: any) => {
    expect(args.customer_code).toBe(customerCode)
    expect(args.origin).toBe('127.0.0.1')
    return success ? Promise.resolve({ url: url }) : Promise.reject(new Error('fetchSavesDashboardMock Mock Error'))
  })
}

function fetchCurrentSavesMock (vueInstance: VueConstructor<any>, success: boolean, saves: Array<any>, customerCode: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchCurrentSaves').mockImplementation((args: any) => {
    expect(args.customer_code).toBe(customerCode)
    return success ? Promise.resolve({ saves: saves }) : Promise.reject(new Error('fetchCurrentSavesMock Mock Error'))
  })
}

describe('Saves.vue', () => {
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

  // Setup spy functions
  let windowSpy: jest.SpyInstance

  beforeEach(() => {
    windowSpy = jest.spyOn(window, 'window', 'get')
    windowSpy.mockImplementation(() => ({
      location: {
        origin: '127.0.0.1'
      }
    }))
  })

  afterEach(() => {
    windowSpy.mockRestore()
  })

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
    const savesDashSpy = fetchSavesDashboardMock(localVue, true, '', '')
    const currentSavesSpy = fetchCurrentSavesMock(localVue, true, [], '')

    const wrapper = shallowMount(Saves, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    wrapper.destroy()

    savesDashSpy.mockReset()
    currentSavesSpy.mockReset()
  })

  it('mounts a single Saves dashboard', async () => {
    const store = new Vuex.Store(defaultStore)
    const savesDashSpy = fetchSavesDashboardMock(localVue, true, 'https://127.0.0.1', '')
    const currentSavesSpy = fetchCurrentSavesMock(localVue, true, [], '')

    const wrapper = shallowMount(Saves, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    const lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('savesDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    wrapper.destroy()

    savesDashSpy.mockReset()
    currentSavesSpy.mockReset()
  })

  it('mounts a single Saves dashboard - failure', async () => {
    const store = new Vuex.Store(defaultStore)
    const savesDashSpy = fetchSavesDashboardMock(localVue, false, 'https://127.0.0.1', '')
    const currentSavesSpy = fetchCurrentSavesMock(localVue, true, [], '')

    const wrapper = shallowMount(Saves, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    const lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(0)

    wrapper.destroy()

    savesDashSpy.mockReset()
    currentSavesSpy.mockReset()
  })

  it('checks filter change update - !admin', async () => {
    const store = new Vuex.Store(defaultStore)
    const savesDashSpy = fetchSavesDashboardMock(localVue, true, 'https://127.0.0.1', '')
    const currentSavesSpy = fetchCurrentSavesMock(localVue, true, [], '')

    const wrapper = shallowMount(Saves, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    let lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('savesDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    const newFilter = { companyCode: 'test' }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    await localVue.nextTick()
    await localVue.nextTick()

    lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('savesDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '' })

    wrapper.destroy()

    savesDashSpy.mockReset()
    currentSavesSpy.mockReset()
  })

  it('checks filter change update - admin', async () => {
    let savesDashSpy = fetchSavesDashboardMock(localVue, true, 'https://127.0.0.1', '')
    let currentSavesSpy = fetchCurrentSavesMock(localVue, true, [], '')

    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)

    const wrapper = shallowMount(Saves, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    let lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('savesDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.1')

    savesDashSpy = fetchSavesDashboardMock(localVue, true, 'https://127.0.0.2', 'test')
    currentSavesSpy = fetchCurrentSavesMock(localVue, true, [], 'test')

    const newFilter = { companyCode: 'test' }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(2)
    expect(currentSavesSpy).toBeCalledTimes(2)

    await localVue.nextTick()

    lookEmbeds = wrapper.findAll('lookerembed-stub')

    expect(lookEmbeds.length).toBe(1)

    expect(lookEmbeds.at(0).attributes('dashid')).toBe('savesDash')
    expect(lookEmbeds.at(0).attributes('dashurl')).toBe('https://127.0.0.2')

    wrapper.destroy()

    store.commit('UPDATE_IS_ADMIN', false)
    store.commit('UPDATE_CURRENT_FILTER', { company_code: '' })

    savesDashSpy.mockReset()
    currentSavesSpy.mockReset()
  })

  it('checks saves table - empty', async () => {
    const store = new Vuex.Store(defaultStore)
    const savesDashSpy = fetchSavesDashboardMock(localVue, true, 'https://127.0.0.1', '')
    const currentSavesSpy = fetchCurrentSavesMock(localVue, true, [], '')

    const wrapper = shallowMount(Saves, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()
    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    const savesTable = wrapper.find('table')
    expect(savesTable.exists()).toBeTruthy()

    const headers = savesTable.findAll('th span')
    expect(headers.length).toBe(5)
    expect(headers.at(0).text()).toBe('Session ID')
    expect(headers.at(1).text()).toBe('Save Score')
    expect(headers.at(2).text()).toBe('RTT Score')
    expect(headers.at(3).text()).toBe('PL Score')
    expect(headers.at(4).text()).toBe('Duration (Hours)')

    const dataRow = savesTable.findAll('tr td')
    expect(dataRow.length).toBe(1)
    expect(dataRow.at(0).text()).toBe('There are no saves at this time.')

    wrapper.destroy()

    savesDashSpy.mockReset()
    currentSavesSpy.mockReset()
  })

  it('checks saves table - !empty', async () => {
    const store = new Vuex.Store(defaultStore)
    const savesDashSpy = fetchSavesDashboardMock(localVue, true, 'https://127.0.0.1', '')
    const currentSavesSpy = fetchCurrentSavesMock(localVue, true, [
      {
        id: '00000000',
        save_score: 1000,
        rtt_score: 1000,
        pl_score: 1000,
        duration: 1
      },
      {
        id: '00000001',
        save_score: 100,
        rtt_score: 100,
        pl_score: 100,
        duration: 10
      }
    ], '')

    const wrapper = shallowMount(Saves, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()
    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    const savesTable = wrapper.find('table')
    expect(savesTable.exists()).toBeTruthy()

    const headers = savesTable.findAll('th span')
    expect(headers.length).toBe(5)
    expect(headers.at(0).text()).toBe('Session ID')
    expect(headers.at(1).text()).toBe('Save Score')
    expect(headers.at(2).text()).toBe('RTT Score')
    expect(headers.at(3).text()).toBe('PL Score')
    expect(headers.at(4).text()).toBe('Duration (Hours)')

    const dataRowCols = savesTable.findAll('tr td')
    expect(dataRowCols.length).toBe(10)

    expect(dataRowCols.at(0).text()).toBe('00000000')
    expect(dataRowCols.at(1).text()).toBe('1000')
    expect(dataRowCols.at(2).text()).toBe('1000')
    expect(dataRowCols.at(3).text()).toBe('1000')
    expect(dataRowCols.at(4).text()).toBe('1')
    expect(dataRowCols.at(5).text()).toBe('00000001')
    expect(dataRowCols.at(6).text()).toBe('100')
    expect(dataRowCols.at(7).text()).toBe('100')
    expect(dataRowCols.at(8).text()).toBe('100')
    expect(dataRowCols.at(9).text()).toBe('10')

    wrapper.destroy()

    savesDashSpy.mockReset()
    currentSavesSpy.mockReset()
  })

  it('checks saves table - failure', async () => {
    const store = new Vuex.Store(defaultStore)
    const savesDashSpy = fetchSavesDashboardMock(localVue, true, 'https://127.0.0.1', '')
    const currentSavesSpy = fetchCurrentSavesMock(localVue, false, [], '')

    const wrapper = shallowMount(Saves, { localVue, store, mocks })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()
    await localVue.nextTick()
    await localVue.nextTick()

    expect(savesDashSpy).toBeCalledTimes(1)
    expect(currentSavesSpy).toBeCalledTimes(1)

    const savesTable = wrapper.find('table')
    expect(savesTable.exists()).toBeTruthy()

    const headers = savesTable.findAll('th span')
    expect(headers.length).toBe(5)
    expect(headers.at(0).text()).toBe('Session ID')
    expect(headers.at(1).text()).toBe('Save Score')
    expect(headers.at(2).text()).toBe('RTT Score')
    expect(headers.at(3).text()).toBe('PL Score')
    expect(headers.at(4).text()).toBe('Duration (Hours)')

    const dataRow = savesTable.findAll('tr td')
    expect(dataRow.length).toBe(1)
    expect(dataRow.at(0).text()).toBe('There are no saves at this time.')

    wrapper.destroy()

    savesDashSpy.mockReset()
    currentSavesSpy.mockReset()
  })
})

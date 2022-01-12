import { shallowMount, createLocalVue } from '@vue/test-utils'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import Vuex from 'vuex'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faCircle } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import VueTour from 'vue-tour'
import { VueConstructor } from 'vue/types/umd'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import { MAX_RETRIES } from '@/components/types/Constants'

function topSessionsMock (vueInstance: VueConstructor<any>, success: boolean, sessions: Array<any>, customerCode: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchTopSessions').mockImplementation((args: any) => {
    expect(args.company_code).toBe(customerCode)

    return success ? Promise.resolve(
      {
        sessions: sessions
      }
    ) : Promise.reject()
  })
}

describe('SessionsWorkspace.vue', () => {
  jest.useFakeTimers()

  const localVue = createLocalVue()
  const defaultStore = {
    state: {
      allBuyers: [],
      filter: {
        companyCode: '',
        dataRange: DateFilterType.CURRENT_MONTH
      },
      killLoops: false,
      isAnonymous: false,
      isAdmin: false,
      isTour: false
    },
    getters: {
      allBuyers: (state: any) => state.allBuyers,
      currentPage: (state: any) => state.currentPage,
      currentFilter: (state: any) => state.filter,
      isAnonymous: (state: any) => state.isAnonymous,
      isAdmin: (state: any) => state.isAdmin,
      isTour: (state: any) => state.isTour,
      killLoops: (state: any) => state.killLoops
    },
    actions: {
      toggleKillLoops ({ commit }: any, killLoops: boolean) {
        commit('TOGGLE_KILL_LOOPS', killLoops)
      }
    },
    mutations: {
      UPDATE_ALL_BUYERS (state: any, allBuyers: Array<any>) {
        state.allBuyers = allBuyers
      },
      UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
        state.filter = newFilter
      },
      TOGGLE_KILL_LOOPS (state: any, killLoops: boolean) {
        state.killLoops = killLoops
      },
      UPDATE_IS_ANONYMOUS (state: any, isAnonymous: boolean) {
        state.isAnonymous = isAnonymous
      },
      UPDATE_IS_TOUR (state: any, isTour: boolean) {
        state.isTour = isTour
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      }
    }
  }


  const ICONS = [
    faCircle
  ]

  const stubs = [
    'router-link'
  ]

  library.add(...ICONS)

  localVue.component('font-awesome-icon', FontAwesomeIcon)

  localVue.use(JSONRPCPlugin)
  localVue.use(VueTour)

  it('mounts the sessions workspace table successfully', () => {
    const spy = topSessionsMock(localVue, true, [], '')
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })
    expect(wrapper.exists()).toBeTruthy()

    spy.mockReset()

    wrapper.destroy()
  })

  it('check sessions table loads - anonymous', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_ANONYMOUS', true)

    const session = {
      on_network_next: true,
      id: '123456789',
      user_hash: '00000000',
      datacenter_alias: 'local_alias',
      location: {
        isp: 'local'
      },
      direct_rtt: 120,
      next_rtt: 20,
      delta_rtt: 100
    }
    const spy = topSessionsMock(localVue, true, [session], '')

    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBeFalsy()

    const headers = wrapper.findAll('thead tr th span')
    expect(headers.length).toBe(7)

    expect(headers.at(0).text()).toBe('')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('ISP')
    expect(headers.at(3).text()).toBe('Datacenter')
    expect(headers.at(4).text()).toBe('Direct RTT')
    expect(headers.at(5).text()).toBe('Next RTT')
    expect(headers.at(6).text()).toBe('Improvement')

    const dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7)

    expect(dataRows.at(0).find('font-awesome-icon-stub').classes('text-success')).toBeTruthy()
    expect(dataRows.at(1).text()).toBe(session.id.toString())
    expect(dataRows.at(2).text()).toBe(session.location.isp)
    expect(dataRows.at(3).text()).toBe(session.datacenter_alias)
    expect(dataRows.at(4).text()).toBe(session.direct_rtt.toFixed(2).toString())
    expect(dataRows.at(5).text()).toBe(session.next_rtt.toFixed(2).toString())
    expect(dataRows.at(6).text()).toBe(session.delta_rtt.toFixed(2).toString())
    expect(dataRows.at(6).find('span').classes().includes('text-success')).toBeTruthy()

    store.commit('UPDATE_IS_ANONYMOUS', false)
    spy.mockReset()

    wrapper.destroy()
  })

  it('check sessions table loads', async () => {
    const store = new Vuex.Store(defaultStore)
    const session = {
      on_network_next: true,
      id: '123456789',
      user_hash: '00000000',
      datacenter_alias: 'local_alias',
      location: {
        isp: 'local'
      },
      direct_rtt: 120,
      next_rtt: 20,
      delta_rtt: 100
    }
    const spy = topSessionsMock(localVue, true, [session], '')

    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBeFalsy()

    const headers = wrapper.findAll('thead tr th span')
    expect(headers.length).toBe(8)

    expect(headers.at(0).text()).toBe('')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('User Hash')
    expect(headers.at(3).text()).toBe('ISP')
    expect(headers.at(4).text()).toBe('Datacenter')
    expect(headers.at(5).text()).toBe('Direct RTT')
    expect(headers.at(6).text()).toBe('Next RTT')
    expect(headers.at(7).text()).toBe('Improvement')

    const dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(8)

    expect(dataRows.at(0).find('font-awesome-icon-stub').classes('text-success')).toBeTruthy()
    expect(dataRows.at(1).text()).toBe(session.id.toString())
    expect(dataRows.at(2).text()).toBe(session.user_hash.toString())
    expect(dataRows.at(3).text()).toBe(session.location.isp)
    expect(dataRows.at(4).text()).toBe(session.datacenter_alias)
    expect(dataRows.at(5).text()).toBe(session.direct_rtt.toFixed(2).toString())
    expect(dataRows.at(6).text()).toBe(session.next_rtt.toFixed(2).toString())
    expect(dataRows.at(7).text()).toBe(session.delta_rtt.toFixed(2).toString())
    expect(dataRows.at(7).find('span').classes().includes('text-success')).toBeTruthy()

    spy.mockReset()

    wrapper.destroy()
  })

  it('check sessions table loads - admin', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_ADMIN', true)
    store.commit('UPDATE_ALL_BUYERS', [
      {
        id: '11111111',
        company_name: 'Test Company'
      }
    ])

    const session = {
      on_network_next: true,
      id: '123456789',
      customer_id: '11111111',
      user_hash: '00000000',
      datacenter_alias: 'local_alias',
      location: {
        isp: 'local'
      },
      direct_rtt: 120,
      next_rtt: 20,
      delta_rtt: 100
    }
    const spy = topSessionsMock(localVue, true, [session], '')

    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBeFalsy()

    const headers = wrapper.findAll('thead tr th span')
    expect(headers.length).toBe(9)

    expect(headers.at(0).text()).toBe('')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('User Hash')
    expect(headers.at(3).text()).toBe('ISP')
    expect(headers.at(4).text()).toBe('Customer')
    expect(headers.at(5).text()).toBe('Datacenter')
    expect(headers.at(6).text()).toBe('Direct RTT')
    expect(headers.at(7).text()).toBe('Next RTT')
    expect(headers.at(8).text()).toBe('Improvement')

    const dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(9)

    expect(dataRows.at(0).find('font-awesome-icon-stub').classes('text-success')).toBeTruthy()
    expect(dataRows.at(1).text()).toBe(session.id.toString())
    expect(dataRows.at(2).text()).toBe(session.user_hash.toString())
    expect(dataRows.at(3).text()).toBe(session.location.isp)
    expect(dataRows.at(4).text()).toBe('Test Company')
    expect(dataRows.at(5).text()).toBe(session.datacenter_alias)
    expect(dataRows.at(6).text()).toBe(session.direct_rtt.toFixed(2).toString())
    expect(dataRows.at(7).text()).toBe(session.next_rtt.toFixed(2).toString())
    expect(dataRows.at(8).text()).toBe(session.delta_rtt.toFixed(2).toString())
    expect(dataRows.at(8).find('span').classes().includes('text-success')).toBeTruthy()

    store.commit('UPDATE_IS_ADMIN', false)
    spy.mockReset()

    wrapper.destroy()
  })

  it('check sessions table loads with different values', async () => {
    const store = new Vuex.Store(defaultStore)
    const session = {
      on_network_next: false,
      id: '123456789',
      user_hash: '00000000',
      datacenter_alias: 'local_alias',
      location: {
        isp: 'local'
      },
      direct_rtt: 0,
      next_rtt: 0,
      delta_rtt: 0
    }
    let spy = topSessionsMock(localVue, true, [session], '')

    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBeFalsy()

    const headers = wrapper.findAll('thead tr th span')
    expect(headers.length).toBe(8)

    expect(headers.at(0).text()).toBe('')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('User Hash')
    expect(headers.at(3).text()).toBe('ISP')
    expect(headers.at(4).text()).toBe('Datacenter')
    expect(headers.at(5).text()).toBe('Direct RTT')
    expect(headers.at(6).text()).toBe('Next RTT')
    expect(headers.at(7).text()).toBe('Improvement')

    let dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(8)

    expect(dataRows.at(0).find('font-awesome-icon-stub').classes('text-primary')).toBeTruthy()
    expect(dataRows.at(1).text()).toBe(session.id.toString())
    expect(dataRows.at(2).text()).toBe(session.user_hash.toString())
    expect(dataRows.at(3).text()).toBe(session.location.isp)
    expect(dataRows.at(4).text()).toBe(session.datacenter_alias)
    expect(dataRows.at(5).text()).toBe('-')
    expect(dataRows.at(6).text()).toBe('-')
    expect(dataRows.at(7).text()).toBe('-')

    session.delta_rtt = 1
    session.next_rtt = 10
    session.direct_rtt = 11
    session.on_network_next = true
    spy = topSessionsMock(localVue, true, [session], '')

    jest.advanceTimersByTime(10000)

    expect(spy).toBeCalledTimes(2)

    await localVue.nextTick()

    dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(8)

    expect(dataRows.at(0).find('font-awesome-icon-stub').classes('text-success')).toBeTruthy()
    expect(dataRows.at(1).text()).toBe(session.id.toString())
    expect(dataRows.at(2).text()).toBe(session.user_hash.toString())
    expect(dataRows.at(3).text()).toBe(session.location.isp)
    expect(dataRows.at(4).text()).toBe(session.datacenter_alias)
    expect(dataRows.at(5).text()).toBe(session.direct_rtt.toFixed(2).toString())
    expect(dataRows.at(6).text()).toBe(session.next_rtt.toFixed(2).toString())
    expect(dataRows.at(7).text()).toBe(session.delta_rtt.toFixed(2).toString())
    expect(dataRows.at(7).find('span').classes().includes('text-danger')).toBeTruthy()

    session.delta_rtt = 3
    session.next_rtt = 10
    session.direct_rtt = 13
    spy = topSessionsMock(localVue, true, [session], '')

    jest.advanceTimersByTime(10000)

    expect(spy).toBeCalledTimes(3)

    await localVue.nextTick()

    dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(8)

    expect(dataRows.at(0).find('font-awesome-icon-stub').classes('text-success')).toBeTruthy()
    expect(dataRows.at(1).text()).toBe(session.id.toString())
    expect(dataRows.at(2).text()).toBe(session.user_hash.toString())
    expect(dataRows.at(3).text()).toBe(session.location.isp)
    expect(dataRows.at(4).text()).toBe(session.datacenter_alias)
    expect(dataRows.at(5).text()).toBe(session.direct_rtt.toFixed(2).toString())
    expect(dataRows.at(6).text()).toBe(session.next_rtt.toFixed(2).toString())
    expect(dataRows.at(7).text()).toBe(session.delta_rtt.toFixed(2).toString())
    expect(dataRows.at(7).find('span').classes().includes('text-warning')).toBeTruthy()

    session.delta_rtt = 0
    session.next_rtt = 14
    session.direct_rtt = 10
    spy = topSessionsMock(localVue, true, [session], '')

    jest.advanceTimersByTime(10000)

    expect(spy).toBeCalledTimes(4)

    await localVue.nextTick()

    dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(8)

    expect(dataRows.at(0).find('font-awesome-icon-stub').classes('text-success')).toBeTruthy()
    expect(dataRows.at(1).text()).toBe(session.id.toString())
    expect(dataRows.at(2).text()).toBe(session.user_hash.toString())
    expect(dataRows.at(3).text()).toBe(session.location.isp)
    expect(dataRows.at(4).text()).toBe(session.datacenter_alias)
    expect(dataRows.at(5).text()).toBe(session.direct_rtt.toFixed(2).toString())
    expect(dataRows.at(6).text()).toBe(session.next_rtt.toFixed(2).toString())
    expect(dataRows.at(7).text()).toBe('')

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks kill loops', async () => {
    const store = new Vuex.Store(defaultStore)
    const session = {
      on_network_next: false,
      id: '123456789',
      user_hash: '00000000',
      datacenter_alias: 'local_alias',
      location: {
        isp: 'local'
      },
      direct_rtt: 0,
      next_rtt: 0,
      delta_rtt: 0
    }
    let spy = topSessionsMock(localVue, true, [session], '')

    let wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })
    expect(wrapper.exists()).toBeTruthy()

    store.commit('TOGGLE_KILL_LOOPS', true)

    await localVue.nextTick()

    expect(spy).toBeCalled()

    wrapper.destroy()

    spy.mockReset()

    spy = topSessionsMock(localVue, true, [session], '')

    wrapper = shallowMount(SessionsWorkspace, { localVue, store })

    await localVue.nextTick()

    expect(spy).not.toBeCalled()

    store.commit('TOGGLE_KILL_LOOPS', false)

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks filter change update', async () => {
    const store = new Vuex.Store(defaultStore)
    const session = {
      on_network_next: false,
      id: '123456789',
      user_hash: '00000000',
      datacenter_alias: 'local_alias',
      location: {
        isp: 'local'
      },
      direct_rtt: 0,
      next_rtt: 0,
      delta_rtt: 0
    }
    let spy = topSessionsMock(localVue, true, [session], '')

    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    spy = topSessionsMock(localVue, true, [session], 'test')

    const newFilter: Filter = { companyCode: 'test', dateRange: DateFilterType.CURRENT_MONTH }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(2)

    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '', dateRange: DateFilterType.CURRENT_MONTH })

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks failed api call', async () => {
    const store = new Vuex.Store(defaultStore)

    const spy = topSessionsMock(localVue, false, [], '')

    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    await localVue.nextTick()

    expect(store.getters.killLoops).toBeFalsy()

    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    let retryCount = wrapper.vm.$data.retryCount
    expect(retryCount).toBe(1)

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBeFalsy()

    const headers = wrapper.findAll('thead tr th span')
    expect(headers.length).toBe(8)

    expect(headers.at(0).text()).toBe('')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('User Hash')
    expect(headers.at(3).text()).toBe('ISP')
    expect(headers.at(4).text()).toBe('Datacenter')
    expect(headers.at(5).text()).toBe('Direct RTT')
    expect(headers.at(6).text()).toBe('Next RTT')
    expect(headers.at(7).text()).toBe('Improvement')

    const dataElement = wrapper.find('tbody tr td')
    expect(dataElement.exists()).toBeTruthy()
    expect(dataElement.text()).toBe('There are no top sessions at this time.')

    for (let i = 2; i <= MAX_RETRIES; i++) {
      jest.advanceTimersByTime(3000 * retryCount)

      await localVue.nextTick()

      expect(spy).toBeCalledTimes(i)

      await localVue.nextTick()

      retryCount = wrapper.vm.$data.retryCount
      expect(retryCount).toBe(i)
    }

    expect(store.getters.killLoops).toBeTruthy()

    wrapper.destroy()
  })
})

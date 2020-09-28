import { shallowMount, createLocalVue } from '@vue/test-utils'
import SessionsWorkspace from '@/workspaces/SessionsWorkspace.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import Vuex from 'vuex'
import { waitFor } from './utils'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faCircle } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

describe('UserSessions.vue no sessions', () => {
  const localVue = createLocalVue()
  const store = new Vuex.Store({
    state: {
      filter: {
        buyerID: ''
      },
      showTable: false
    },
    getters: {
      currentFilter: (state: any) => state.filter,
      isAnonymous: () => true
    },
    mutations: {
      TOGGLE_SESSION_TABLE (state: any, showTable: boolean) {
        state.showTable = showTable
      }
    }
  })

  const ICONS = [
    faCircle
  ]

  const stubs = [
    'router-link'
  ]

  library.add(...ICONS)

  localVue.component('font-awesome-icon', FontAwesomeIcon)

  localVue.use(JSONRPCPlugin)

  const spy = jest.spyOn(localVue.prototype.$apiService, 'fetchTopSessions').mockImplementation(() => {
    return Promise.resolve({
      sessions: [
        {
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
      ]
    })
  })

  it('mounts the sessions workspace table successfully', () => {
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })

  it('check sessions table loads for all filter with high delta rtt', async () => {
    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalled()

    // Wait for the api call to come back and update the internal state
    await waitFor(wrapper, 'table')

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBe(false)
    expect(wrapper.findAll('thead tr th span').length).toBe(7)

    const dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7)

    expect(dataRows.at(0).text()).toBe('')
    expect(dataRows.at(1).text()).toBe('123456789')
    expect(dataRows.at(2).text()).toBe('local')
    expect(dataRows.at(3).text()).toBe('local_alias')
    expect(dataRows.at(4).text()).toBe('120.00')
    expect(dataRows.at(5).text()).toBe('20.00')
    expect(dataRows.at(6).text()).toBe('100.00')
    expect(dataRows.at(6).find('span').classes().includes('text-success')).toBe(true)

    wrapper.destroy()
  })

  it('check sessions table loads for all filter with medium delta rtt', async () => {
    const spy = jest.spyOn(localVue.prototype.$apiService, 'fetchTopSessions').mockImplementationOnce(() => {
      return Promise.resolve({
        sessions: [
          {
            on_network_next: true,
            id: '123456789',
            user_hash: '00000000',
            datacenter_alias: 'local_alias',
            location: {
              isp: 'local'
            },
            direct_rtt: 22,
            next_rtt: 20,
            delta_rtt: 2
          }
        ]
      })
    })
    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalled()

    // Wait for the api call to come back and update the internal state
    await waitFor(wrapper, 'table')

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBe(false)
    expect(wrapper.findAll('thead tr th span').length).toBe(7)

    const dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7)

    expect(dataRows.at(0).text()).toBe('')
    expect(dataRows.at(1).text()).toBe('123456789')
    expect(dataRows.at(2).text()).toBe('local')
    expect(dataRows.at(3).text()).toBe('local_alias')
    expect(dataRows.at(4).text()).toBe('22.00')
    expect(dataRows.at(5).text()).toBe('20.00')
    expect(dataRows.at(6).text()).toBe('2.00')
    expect(dataRows.at(6).find('span').classes().includes('text-warning')).toBe(true)

    wrapper.destroy()
  })

  it('check sessions table loads for all filter with low delta rtt', async () => {
    const spy = jest.spyOn(localVue.prototype.$apiService, 'fetchTopSessions').mockImplementationOnce(() => {
      return Promise.resolve({
        sessions: [
          {
            on_network_next: true,
            id: '123456789',
            user_hash: '00000000',
            datacenter_alias: 'local_alias',
            location: {
              isp: 'local'
            },
            direct_rtt: 22,
            next_rtt: 21,
            delta_rtt: 1
          }
        ]
      })
    })
    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalled()

    // Wait for the api call to come back and update the internal state
    await waitFor(wrapper, 'table')

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBe(false)
    expect(wrapper.findAll('thead tr th span').length).toBe(7)

    const dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7)

    expect(dataRows.at(0).text()).toBe('')
    expect(dataRows.at(1).text()).toBe('123456789')
    expect(dataRows.at(2).text()).toBe('local')
    expect(dataRows.at(3).text()).toBe('local_alias')
    expect(dataRows.at(4).text()).toBe('22.00')
    expect(dataRows.at(5).text()).toBe('21.00')
    expect(dataRows.at(6).text()).toBe('1.00')
    expect(dataRows.at(6).find('span').classes().includes('text-danger')).toBe(true)

    wrapper.destroy()
  })

  it('check sessions table loads for all filter with multiple sessions', async () => {
    const sessions = [
      {
        on_network_next: true,
        id: '123456789',
        user_hash: '00000000',
        datacenter_alias: 'local_alias',
        location: {
          isp: 'local'
        },
        direct_rtt: 22,
        next_rtt: 21,
        delta_rtt: 1
      },
      {
        on_network_next: false,
        id: '1234567891011',
        user_hash: '0000000000001',
        datacenter_alias: 'local_alias',
        location: {
          isp: 'local'
        },
        direct_rtt: 100,
        next_rtt: 100,
        delta_rtt: 0
      },
    ]
    const spy = jest.spyOn(localVue.prototype.$apiService, 'fetchTopSessions').mockImplementationOnce(() => {
      return Promise.resolve({ sessions: sessions })
    })
    // Mount the component
    const wrapper = shallowMount(SessionsWorkspace, { localVue, store, stubs })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalled()

    // Wait for the api call to come back and update the internal state
    await waitFor(wrapper, 'table')

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    expect(wrapper.find('#session-spinner').exists()).toBe(false)
    expect(wrapper.findAll('thead tr th span').length).toBe(7)

    const dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7 * sessions.length)

    expect(dataRows.at(0).text()).toBe('')
    expect(dataRows.at(0).find('font-awesome-icon-stub').classes().includes('text-success')).toBe(true)
    expect(dataRows.at(1).text()).toBe('123456789')
    expect(dataRows.at(2).text()).toBe('local')
    expect(dataRows.at(3).text()).toBe('local_alias')
    expect(dataRows.at(4).text()).toBe('22.00')
    expect(dataRows.at(5).text()).toBe('21.00')
    expect(dataRows.at(6).text()).toBe('1.00')
    expect(dataRows.at(6).find('span').classes().includes('text-danger')).toBe(true)

    expect(dataRows.at(7).text()).toBe('')
    expect(dataRows.at(7).find('font-awesome-icon-stub').classes().includes('text-primary')).toBe(true)
    expect(dataRows.at(8).text()).toBe('1234567891011')
    expect(dataRows.at(9).text()).toBe('local')
    expect(dataRows.at(10).text()).toBe('local_alias')
    expect(dataRows.at(11).text()).toBe('100.00')
    expect(dataRows.at(12).text()).toBe('100.00')
    expect(dataRows.at(13).text()).toBe('-')

    wrapper.destroy()
  })
})

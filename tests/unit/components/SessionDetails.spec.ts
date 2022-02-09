import { shallowMount, createLocalVue } from '@vue/test-utils'
import SessionDetails from '@/components/SessionDetails.vue'
import Vuex from 'vuex'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { VueConstructor } from 'vue/types/umd'

function fetchSessionDetailsMock (localVue: VueConstructor<any>, success: boolean, meta: any, slices: Array<any>, sessionID: string) {
  return jest.spyOn(localVue.prototype.$apiService, 'fetchSessionDetails').mockImplementation((args: any) => {
    expect(args.session_id).toBe(sessionID)
    return success ? Promise.resolve({
      meta: meta,
      slices: slices
    }) : Promise.reject(new Error('fetchSessionDetailsMock Error'))
  })
}

describe('SessionDetails.vue', () => {
  jest.useFakeTimers()
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const $route = {
    path: '/user-tool',
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

  const stubs = [
    'router-link',
    'v-tour'
  ]

  // Init the store instance
  const defaultStore = {
    state: {
      allBuyers: [],
      userProfile: newDefaultProfile(),
      killLoops: false,
      isAnonymous: true,
      isAnonymousPlus: false,
      isAdmin: false
    },
    getters: {
      allBuyers: (state: any) => state.allBuyers,
      userProfile: (state: any) => state.userProfile,
      killLoops: (state: any) => state.killLoops,
      isAnonymous: (state: any) => state.isAnonymous,
      isAnonymousPlus: (state: any) => state.isAnonymousPlus,
      isAdmin: (state: any) => state.isAdmin
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
      TOGGLE_KILL_LOOPS (state: any, killLoops: boolean) {
        state.killLoops = killLoops
      },
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      },
      UPDATE_IS_ANONYMOUS_PLUS (state: any, isAnonymousPlus: boolean) {
        state.isAnonymousPlus = isAnonymousPlus
      },
      UPDATE_IS_ANONYMOUS (state: any, isAnonymous: boolean) {
        state.isAnonymous = isAnonymous
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      }
    }
  }

  // Run bare minimum mount test
  it('mounts the component successfully', () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(SessionDetails, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    wrapper.destroy()
  })

  // TODO: The meta data panel could probably be its own component and the tests can be pulled out into their own suite
  it('checks meta data fields - anonymous', async () => {
    const spy = fetchSessionDetailsMock(localVue, true, {
      nearby_relays: [],
      datacenter_alias: 'local',
      location: {
        isp: 'local'
      },
      user_hash: '00000000',
      client_addr: '127.0.0.1',
      platform: 'Linux',
      sdk: '4.0.16',
      connection: 'Wired',
      hops: [],
      server_addr: '127.0.0.1'
    }, [], '00000000')
    const store = new Vuex.Store(defaultStore)

    $route.params.pathMatch = '00000000'

    const wrapper = shallowMount(SessionDetails, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    const metaPanel = wrapper.find('#meta-panel')
    expect(metaPanel.exists()).toBeTruthy()

    const metaTitles = wrapper.findAll('dt')
    expect(metaTitles.length).toBe(5)

    expect(metaTitles.at(0).text()).toBe('Datacenter')
    expect(metaTitles.at(1).text()).toBe('ISP')
    expect(metaTitles.at(2).text()).toBe('Platform')
    expect(metaTitles.at(3).text()).toBe('SDK Version')
    expect(metaTitles.at(4).text()).toBe('Connection Type')

    const metaData = wrapper.findAll('dd')
    expect(metaData.length).toBe(5)

    expect(metaData.at(0).text()).toBe('local')
    expect(metaData.at(1).text()).toBe('local')
    expect(metaData.at(2).text()).toBe('Linux')
    expect(metaData.at(3).text()).toBe('4.0.16')
    expect(metaData.at(4).text()).toBe('Wired')

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks meta data fields - !admin - !sameBuyer', async () => {
    const spy = fetchSessionDetailsMock(localVue, true, {
      nearby_relays: [],
      datacenter_alias: 'local',
      location: {
        isp: 'local'
      },
      user_hash: '00000000',
      client_addr: '127.0.0.1',
      platform: 'Linux',
      sdk: '4.0.16',
      connection: 'Wired',
      hops: [],
      server_addr: '127.0.0.1',
      customer_id: '00000000'
    }, [], '00000000')
    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_ANONYMOUS', false)

    $route.params.pathMatch = '00000000'

    const wrapper = shallowMount(SessionDetails, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    const metaPanel = wrapper.find('#meta-panel')
    expect(metaPanel.exists()).toBeTruthy()

    const metaTitles = wrapper.findAll('dt')
    expect(metaTitles.length).toBe(7)

    expect(metaTitles.at(0).text()).toBe('Datacenter')
    expect(metaTitles.at(1).text()).toBe('ISP')
    expect(metaTitles.at(2).text()).toBe('User Hash')
    expect(metaTitles.at(3).text()).toBe('Platform')
    expect(metaTitles.at(4).text()).toBe('Customer')
    expect(metaTitles.at(5).text()).toBe('SDK Version')
    expect(metaTitles.at(6).text()).toBe('Connection Type')

    const metaData = wrapper.findAll('dd')
    expect(metaData.length).toBe(7)

    expect(metaData.at(0).text()).toBe('local')
    expect(metaData.at(1).text()).toBe('local')
    expect(metaData.at(2).text()).toBe('00000000')
    expect(metaData.at(3).text()).toBe('Linux')
    expect(metaData.at(4).text()).toBe('Private')
    expect(metaData.at(5).text()).toBe('4.0.16')
    expect(metaData.at(6).text()).toBe('Wired')

    store.commit('UPDATE_IS_ANONYMOUS', true)

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks meta data fields - !admin - sameBuyer', async () => {
    const spy = fetchSessionDetailsMock(localVue, true, {
      nearby_relays: [],
      datacenter_alias: 'local',
      location: {
        isp: 'local'
      },
      user_hash: '00000000',
      client_addr: '127.0.0.1',
      platform: 'Linux',
      sdk: '4.0.16',
      connection: 'Wired',
      hops: [],
      server_addr: '127.0.0.1',
      customer_id: '00000000'
    }, [], '00000000')

    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_ALL_BUYERS', [
      {
        id: '00000000',
        company_code: 'test',
        company_name: 'Test Company',
        is_live: false
      }
    ])

    const profile = newDefaultProfile()
    profile.buyerID = '00000000'
    profile.companyCode = 'test'
    profile.companyName = 'Test Company'

    store.commit('UPDATE_USER_PROFILE', profile)
    store.commit('UPDATE_IS_ANONYMOUS', false)

    $route.params.pathMatch = '00000000'

    const wrapper = shallowMount(SessionDetails, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    const metaPanel = wrapper.find('#meta-panel')
    expect(metaPanel.exists()).toBeTruthy()

    const metaTitles = wrapper.findAll('dt')
    expect(metaTitles.length).toBe(8)

    expect(metaTitles.at(0).text()).toBe('Datacenter')
    expect(metaTitles.at(1).text()).toBe('ISP')
    expect(metaTitles.at(2).text()).toBe('User Hash')
    expect(metaTitles.at(3).text()).toBe('IP Address')
    expect(metaTitles.at(4).text()).toBe('Platform')
    expect(metaTitles.at(5).text()).toBe('Customer')
    expect(metaTitles.at(6).text()).toBe('SDK Version')
    expect(metaTitles.at(7).text()).toBe('Connection Type')

    const metaData = wrapper.findAll('dd')
    expect(metaData.length).toBe(8)

    expect(metaData.at(0).text()).toBe('local')
    expect(metaData.at(1).text()).toBe('local')
    expect(metaData.at(2).text()).toBe('00000000')
    expect(metaData.at(3).text()).toBe('127.0.0.1')
    expect(metaData.at(4).text()).toBe('Linux')
    expect(metaData.at(5).text()).toBe('Test Company')
    expect(metaData.at(6).text()).toBe('4.0.16')
    expect(metaData.at(7).text()).toBe('Wired')

    store.commit('UPDATE_ALL_BUYERS', [])
    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    store.commit('UPDATE_IS_ANONYMOUS', true)

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks meta data fields - admin - !live', async () => {
    const spy = fetchSessionDetailsMock(localVue, true, {
      nearby_relays: [],
      datacenter_alias: 'local',
      location: {
        isp: 'local'
      },
      user_hash: '00000000',
      client_addr: '127.0.0.1',
      platform: 'Linux',
      sdk: '4.0.16',
      connection: 'Wired',
      hops: [],
      server_addr: '127.0.0.1',
      customer_id: '00000000'
    }, [], '00000000')

    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_ALL_BUYERS', [
      {
        id: '00000000',
        company_code: 'test',
        company_name: 'Test Company',
        is_live: false
      }
    ])

    store.commit('UPDATE_IS_ADMIN', true)
    store.commit('UPDATE_IS_ANONYMOUS', false)

    $route.params.pathMatch = '00000000'

    const wrapper = shallowMount(SessionDetails, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    const metaPanel = wrapper.find('#meta-panel')
    expect(metaPanel.exists()).toBeTruthy()

    const metaTitles = wrapper.findAll('dt')
    expect(metaTitles.length).toBe(9)

    expect(metaTitles.at(0).text()).toBe('Datacenter')
    expect(metaTitles.at(1).text()).toBe('ISP')
    expect(metaTitles.at(2).text()).toBe('User Hash')
    expect(metaTitles.at(3).text()).toBe('IP Address')
    expect(metaTitles.at(4).text()).toBe('Platform')
    expect(metaTitles.at(5).text()).toBe('Customer')
    expect(metaTitles.at(6).text()).toBe('SDK Version')
    expect(metaTitles.at(7).text()).toBe('Connection Type')
    expect(metaTitles.at(8).text()).toBe('Route')

    const metaData = wrapper.findAll('dd')
    expect(metaData.length).toBe(9)

    expect(metaData.at(0).text()).toBe('local')
    expect(metaData.at(1).text()).toBe('local')
    expect(metaData.at(2).text()).toBe('00000000')
    expect(metaData.at(3).text()).toBe('127.0.0.1')
    expect(metaData.at(4).text()).toBe('Linux')
    expect(metaData.at(5).text()).toBe('Test Company')
    expect(metaData.at(6).text()).toBe('4.0.16')
    expect(metaData.at(7).text()).toBe('Wired')
    expect(metaData.at(8).text()).toBe('Customer is not live')

    store.commit('UPDATE_ALL_BUYERS', [])
    store.commit('UPDATE_IS_ANONYMOUS', true)
    store.commit('UPDATE_IS_ADMIN', false)

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks meta data fields - admin - live - no near relays', async () => {
    const spy = fetchSessionDetailsMock(localVue, true, {
      nearby_relays: [],
      datacenter_alias: 'local',
      location: {
        isp: 'local'
      },
      user_hash: '00000000',
      client_addr: '127.0.0.1',
      platform: 'Linux',
      sdk: '4.0.16',
      connection: 'Wired',
      hops: [],
      server_addr: '127.0.0.1',
      customer_id: '00000000'
    }, [], '00000000')

    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_ALL_BUYERS', [
      {
        id: '00000000',
        company_code: 'test',
        company_name: 'Test Company',
        is_live: true
      }
    ])

    store.commit('UPDATE_IS_ADMIN', true)
    store.commit('UPDATE_IS_ANONYMOUS', false)

    $route.params.pathMatch = '00000000'

    const wrapper = shallowMount(SessionDetails, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    const metaPanel = wrapper.find('#meta-panel')
    expect(metaPanel.exists()).toBeTruthy()

    const metaTitles = wrapper.findAll('dt')
    expect(metaTitles.length).toBe(9)

    expect(metaTitles.at(0).text()).toBe('Datacenter')
    expect(metaTitles.at(1).text()).toBe('ISP')
    expect(metaTitles.at(2).text()).toBe('User Hash')
    expect(metaTitles.at(3).text()).toBe('IP Address')
    expect(metaTitles.at(4).text()).toBe('Platform')
    expect(metaTitles.at(5).text()).toBe('Customer')
    expect(metaTitles.at(6).text()).toBe('SDK Version')
    expect(metaTitles.at(7).text()).toBe('Connection Type')
    expect(metaTitles.at(8).text()).toBe('Route')

    const metaData = wrapper.findAll('dd')
    expect(metaData.length).toBe(9)

    expect(metaData.at(0).text()).toBe('local')
    expect(metaData.at(1).text()).toBe('local')
    expect(metaData.at(2).text()).toBe('00000000')
    expect(metaData.at(3).text()).toBe('127.0.0.1')
    expect(metaData.at(4).text()).toBe('Linux')
    expect(metaData.at(5).text()).toBe('Test Company')
    expect(metaData.at(6).text()).toBe('4.0.16')
    expect(metaData.at(7).text()).toBe('Wired')
    expect(metaData.at(8).text()).toBe('No Near Relays')

    store.commit('UPDATE_ALL_BUYERS', [])
    store.commit('UPDATE_IS_ANONYMOUS', true)
    store.commit('UPDATE_IS_ADMIN', false)

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks meta data fields - admin - live', async () => {
    const spy = fetchSessionDetailsMock(localVue, true, {
      nearby_relays: [
        {
          id: '00000000',
          name: 'local.1',
          client_stats: {
            jitter: 100,
            packet_loss: 100,
            rtt: 100,
          }
        },
        {
          id: '00000001',
          name: 'local.2',
          client_stats: {
            jitter: 100,
            packet_loss: 100,
            rtt: 100,
          }
        },
        {
          id: '00000002',
          name: 'local.3',
          client_stats: {
            jitter: 100,
            packet_loss: 100,
            rtt: 100,
          }
        },
        {
          id: '00000003',
          name: 'local.4',
          client_stats: {
            jitter: 100,
            packet_loss: 100,
            rtt: 255,
          }
        }
      ],
      datacenter_alias: 'local',
      location: {
        isp: 'local'
      },
      user_hash: '00000000',
      client_addr: '127.0.0.1',
      platform: 'Linux',
      sdk: '4.0.16',
      connection: 'Wired',
      hops: [
        {
          id: '00000000',
          name: 'local.1'
        },
        {
          id: '00000001',
          name: 'local.2'
        }
      ],
      server_addr: '127.0.0.1',
      customer_id: '00000000'
    }, [], '00000000')

    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_ALL_BUYERS', [
      {
        id: '00000000',
        company_code: 'test',
        company_name: 'Test Company',
        is_live: true
      }
    ])

    store.commit('UPDATE_IS_ADMIN', true)
    store.commit('UPDATE_IS_ANONYMOUS', false)

    $route.params.pathMatch = '00000000'

    const wrapper = shallowMount(SessionDetails, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    const metaPanel = wrapper.find('#meta-panel')
    expect(metaPanel.exists()).toBeTruthy()

    const metaTitles = wrapper.findAll('dt')
    expect(metaTitles.length).toBe(10)

    expect(metaTitles.at(0).text()).toBe('Datacenter')
    expect(metaTitles.at(1).text()).toBe('ISP')
    expect(metaTitles.at(2).text()).toBe('User Hash')
    expect(metaTitles.at(3).text()).toBe('IP Address')
    expect(metaTitles.at(4).text()).toBe('Platform')
    expect(metaTitles.at(5).text()).toBe('Customer')
    expect(metaTitles.at(6).text()).toBe('SDK Version')
    expect(metaTitles.at(7).text()).toBe('Connection Type')
    expect(metaTitles.at(8).text()).toBe('Route')

    const metaData = wrapper.findAll('dd')
    expect(metaData.length).toBe(8)

    expect(metaData.at(0).text()).toBe('local')
    expect(metaData.at(1).text()).toBe('local')
    expect(metaData.at(2).text()).toBe('00000000')
    expect(metaData.at(3).text()).toBe('127.0.0.1')
    expect(metaData.at(4).text()).toBe('Linux')
    expect(metaData.at(5).text()).toBe('Test Company')
    expect(metaData.at(6).text()).toBe('4.0.16')
    expect(metaData.at(7).text()).toBe('Wired')

    const routeTable = wrapper.find('#route-table')
    expect(routeTable.exists()).toBeTruthy()

    const routeHeaders = routeTable.findAll('tr th')
    expect(routeHeaders.length).toBe(2)
    expect(routeHeaders.at(0).text()).toBe('Name')
    expect(routeHeaders.at(1).text()).toBe('Role')

    const routeCells = routeTable.findAll('td')
    expect(routeCells.length).toBe(8)
    expect(routeCells.at(0).text()).toBe('127.0.0.1')
    expect(routeCells.at(1).text()).toBe('User (Player)')
    expect(routeCells.at(2).text()).toBe('local.1')
    expect(routeCells.at(3).text()).toBe('Hop 1')
    expect(routeCells.at(4).text()).toBe('local.2')
    expect(routeCells.at(5).text()).toBe('Hop 2')
    expect(routeCells.at(6).text()).toBe('127.0.0.1')
    expect(routeCells.at(7).text()).toBe('Destination Server')

    const nearbyRelaysTable = wrapper.find('#nearby-relays-table')
    expect(nearbyRelaysTable.exists()).toBeTruthy()

    const nearbyRelaysHeaders = nearbyRelaysTable.findAll('tr th')
    expect(nearbyRelaysHeaders.length).toBe(4)
    expect(nearbyRelaysHeaders.at(0).text()).toBe('Name')
    expect(nearbyRelaysHeaders.at(1).text()).toBe('RTT')
    expect(nearbyRelaysHeaders.at(2).text()).toBe('Jitter')
    expect(nearbyRelaysHeaders.at(3).text()).toBe('Packet Loss')

    const nearbyRelaysCells = nearbyRelaysTable.findAll('td')
    expect(nearbyRelaysCells.length).toBe(16)
    expect(nearbyRelaysCells.at(0).text()).toBe('local.1')
    expect(nearbyRelaysCells.at(1).text()).toBe('100.00')
    expect(nearbyRelaysCells.at(2).text()).toBe('100.00')
    expect(nearbyRelaysCells.at(3).text()).toBe('100.00%')
    expect(nearbyRelaysCells.at(4).text()).toBe('local.2')
    expect(nearbyRelaysCells.at(5).text()).toBe('100.00')
    expect(nearbyRelaysCells.at(6).text()).toBe('100.00')
    expect(nearbyRelaysCells.at(7).text()).toBe('100.00%')
    expect(nearbyRelaysCells.at(8).text()).toBe('local.3')
    expect(nearbyRelaysCells.at(9).text()).toBe('100.00')
    expect(nearbyRelaysCells.at(10).text()).toBe('100.00')
    expect(nearbyRelaysCells.at(11).text()).toBe('100.00%')
    expect(nearbyRelaysCells.at(12).text()).toBe('local.4')
    expect(nearbyRelaysCells.at(13).text()).toBe('-')
    expect(nearbyRelaysCells.at(14).text()).toBe('-')
    expect(nearbyRelaysCells.at(15).text()).toBe('-')

    store.commit('UPDATE_ALL_BUYERS', [])
    store.commit('UPDATE_IS_ANONYMOUS', true)
    store.commit('UPDATE_IS_ADMIN', false)

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks chart legends', async () => {
    const spy = fetchSessionDetailsMock(localVue, true, {
      nearby_relays: [],
      datacenter_alias: 'local',
      location: {
        isp: 'local'
      },
      user_hash: '00000000',
      client_addr: '127.0.0.1',
      platform: 'Linux',
      sdk: '4.0.16',
      connection: 'Wired',
      hops: [],
      server_addr: '127.0.0.1',
      customer_id: '00000000'
    }, [], '00000000')

    const store = new Vuex.Store(defaultStore)

    $route.params.pathMatch = '00000000'

    const wrapper = shallowMount(SessionDetails, { localVue, store, mocks, stubs })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    const latencyCard = wrapper.find('#latency-card')
    expect(latencyCard.exists()).toBeTruthy()

    const latencyCardTitle = latencyCard.find('strong')
    expect(latencyCardTitle.exists()).toBeTruthy()
    expect(latencyCardTitle.text()).toBe('Latency')

    const latencyLegends = latencyCard.findAll('span')
    expect(latencyLegends.length).toBe(2)
    expect(latencyLegends.at(0).text()).toBe('— Network Next')
    expect(latencyLegends.at(1).text()).toBe('— Direct')

    const jitterCard = wrapper.find('#jitter-card')
    expect(jitterCard.exists()).toBeTruthy()

    const jitterCardTitle = jitterCard.find('strong')
    expect(jitterCardTitle.exists()).toBeTruthy()
    expect(jitterCardTitle.text()).toBe('Jitter')

    const jitterLegends = jitterCard.findAll('span')
    expect(jitterLegends.length).toBe(2)
    expect(jitterLegends.at(0).text()).toBe('— Network Next')
    expect(jitterLegends.at(1).text()).toBe('— Direct')

    const plCard = wrapper.find('#pl-card')
    expect(plCard.exists()).toBeTruthy()

    const plCardTitle = plCard.find('strong')
    expect(plCardTitle.exists()).toBeTruthy()
    expect(plCardTitle.text()).toBe('Packet Loss')

    const plLegends = plCard.findAll('span')
    expect(plLegends.length).toBe(2)
    expect(plLegends.at(0).text()).toBe('— Network Next')
    expect(plLegends.at(1).text()).toBe('— Direct')

    const bwCard = wrapper.find('#bw-card')
    expect(bwCard.exists()).toBeTruthy()

    const bwCardTitle = bwCard.find('strong')
    expect(bwCardTitle.exists()).toBeTruthy()
    expect(bwCardTitle.text()).toBe('Bandwidth')

    const bwLegends = bwCard.findAll('span')
    expect(bwLegends.length).toBe(2)
    expect(bwLegends.at(0).text()).toBe('— Up')
    expect(bwLegends.at(1).text()).toBe('— Down')

    spy.mockReset()
    wrapper.destroy()
  })
})

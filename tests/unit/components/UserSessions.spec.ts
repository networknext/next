import { createLocalVue, shallowMount } from '@vue/test-utils'
import UserSessions from '@/components/UserSessions.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { MAX_USER_SESSION_PAGES } from '@/components/types/Constants'
import { FlagPlugin } from '@/plugins/flags'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import Vuex from 'vuex'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { entries, result } from 'lodash'
import { FeatureEnum } from '@/components/types/FeatureTypes'

function fetchUserSessionsMock (vueInstance: VueConstructor<any>, success: boolean, sessions: Array<any>, nextPage: number, userID: string, page: number): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchUserSessions').mockImplementation((args: any) => {
    expect(args.user_id).toBe(userID)
    expect(args.page).toBe(page)

    return success ? Promise.resolve({ sessions: sessions, page: nextPage }) : Promise.reject(new Error('Mock Error'))
  })
}

function fetchLookerUserSessionsMock (vueInstance: VueConstructor<any>, success: boolean, sessions: Array<any>, userID: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchUserSessions').mockImplementation((args: any) => {
    expect(args.user_id).toBe(userID)

    return success ? Promise.resolve({ sessions: sessions }) : Promise.reject(new Error('Mock Error'))
  })
}

describe('UserSessions.vue no sessions', () => {
  const localVue = createLocalVue()
  localVue.use(FlagPlugin, {
    flags: [
      {
        name: FeatureEnum.FEATURE_LOOKER_BIGTABLE_REPLACEMENT,
        description: 'Leverage Looker API for user tool and session tool',
        value: true
      }
    ],
    useAPI: false,
    apiService: {}
  })

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

  localVue.use(JSONRPCPlugin)
  localVue.use(FlagPlugin, {
    flags: [],
    useAPI: false,
    apiService: {}
  })

  const defaultStore = {
    state: {
      filter: {
        companyCode: '',
        dataRange: DateFilterType.LAST_7
      },
      userProfile: newDefaultProfile(),
      isAnonymous: false,
      isAnonymousPlus: false,
      isAdmin: false
    },
    getters: {
      currentPage: (state: any) => state.currentPage,
      currentFilter: (state: any) => state.filter,
      isAnonymous: (state: any) => state.isAnonymous,
      isAnonymousPlus: (state: any) => state.isAnonymousPlus,
      isAdmin: (state: any) => state.isAdmin,
      userProfile: (state: any) => state.userProfile
    },
    actions: {},
    mutations: {
      UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
        state.filter = newFilter
      },
      UPDATE_IS_ANONYMOUS (state: any, isAnonymous: boolean) {
        state.isAnonymous = isAnonymous
      },
      UPDATE_IS_ANONYMOUS_PLUS (state: any, isAnonymousPlus: boolean) {
        state.isAnonymousPlus = isAnonymousPlus
      },
      UPDATE_IS_TOUR (state: any, isTour: boolean) {
        state.isTour = isTour
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      },
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      }
    }
  }

  const stubs = [
    'router-link'
  ]

  it('mounts the user sessions table successfully', () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks, store
    })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('checks default view with no sessions', async () => {
    const store = new Vuex.Store(defaultStore)

    $route.path = '/user-tool/00000000'
    $route.params.pathMatch = '00000000'

    const sessionsSpy = fetchUserSessionsMock(localVue, true, [], MAX_USER_SESSION_PAGES, '00000000', 0)

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    const headers = wrapper.findAll('thead tr th span')
    expect(headers.length).toBe(7)

    expect(headers.at(0).text()).toBe('Date')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('Platform')
    expect(headers.at(3).text()).toBe('Connection Type')
    expect(headers.at(4).text()).toBe('ISP')
    expect(headers.at(5).text()).toBe('Datacenter')
    expect(headers.at(6).text()).toBe('Server Address')

    let dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(1)

    expect(dataRows.at(0).text()).toBe('There are no sessions belonging to this user.')

    const buttons = wrapper.findAll('button')

    expect(buttons.length).toBe(1)

    expect(buttons.at(0).text()).toBe('Refresh Sessions')

    await buttons.at(0).trigger('click')

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(2)

    await localVue.nextTick()

    dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(1)

    expect(dataRows.at(0).text()).toBe('There are no sessions belonging to this user.')

    sessionsSpy.mockReset()

    wrapper.destroy()
  })

  it('checks default view with sessions - 1 page', async () => {
    const store = new Vuex.Store(defaultStore)

    $route.path = '/user-tool/00000000'
    $route.params.pathMatch = '00000000'

    const sessionsSpy = fetchUserSessionsMock(localVue, true, [
      {
        time_stamp: new Date(),
        meta: {
          id: '00000000',
          platform: 'test',
          connection: 'wifi',
          location: {
            isp: 'test'
          },
          datacenter_alias: 'test',
          server_addr: '127.0.0.1'
        }
      }
    ], MAX_USER_SESSION_PAGES, '00000000', 0)

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    const headers = wrapper.findAll('thead tr th span')
    expect(headers.length).toBe(7)

    expect(headers.at(0).text()).toBe('Date')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('Platform')
    expect(headers.at(3).text()).toBe('Connection Type')
    expect(headers.at(4).text()).toBe('ISP')
    expect(headers.at(5).text()).toBe('Datacenter')
    expect(headers.at(6).text()).toBe('Server Address')

    let dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7)

    expect(dataRows.at(0).text()).not.toBe('') // Formatted date based on test time
    expect(dataRows.at(1).text()).toBe('00000000')
    expect(dataRows.at(2).text()).toBe('test')
    expect(dataRows.at(3).text()).toBe('Wi-Fi')
    expect(dataRows.at(4).text()).toBe('test')
    expect(dataRows.at(5).text()).toBe('test')
    expect(dataRows.at(6).text()).toBe('127.0.0.1')

    const buttons = wrapper.findAll('button')

    expect(buttons.length).toBe(1)

    expect(buttons.at(0).text()).toBe('Refresh Sessions')

    await buttons.at(0).trigger('click')

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(2)

    await localVue.nextTick()

    dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7)

    expect(dataRows.at(0).text()).not.toBe('') // Formatted date based on test time
    expect(dataRows.at(1).text()).toBe('00000000')
    expect(dataRows.at(2).text()).toBe('test')
    expect(dataRows.at(3).text()).toBe('Wi-Fi')
    expect(dataRows.at(4).text()).toBe('test')
    expect(dataRows.at(5).text()).toBe('test')
    expect(dataRows.at(6).text()).toBe('127.0.0.1')

    sessionsSpy.mockReset()

    wrapper.destroy()
  })

  it('checks default view with sessions - 2 pages', async () => {
    const store = new Vuex.Store(defaultStore)

    $route.path = '/user-tool/00000000'
    $route.params.pathMatch = '00000000'

    let sessionsSpy = fetchUserSessionsMock(localVue, true, [
      {
        time_stamp: new Date(),
        meta: {
          id: '00000000',
          platform: 'test',
          connection: 'wifi',
          location: {
            isp: 'test'
          },
          datacenter_alias: 'test',
          server_addr: '127.0.0.1'
        }
      }
    ], 1, '00000000', 0)

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    const headers = wrapper.findAll('thead tr th span')
    expect(headers.length).toBe(7)

    expect(headers.at(0).text()).toBe('Date')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('Platform')
    expect(headers.at(3).text()).toBe('Connection Type')
    expect(headers.at(4).text()).toBe('ISP')
    expect(headers.at(5).text()).toBe('Datacenter')
    expect(headers.at(6).text()).toBe('Server Address')

    let dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7)

    expect(dataRows.at(0).text()).not.toBe('') // Formatted date based on test time
    expect(dataRows.at(1).text()).toBe('00000000')
    expect(dataRows.at(2).text()).toBe('test')
    expect(dataRows.at(3).text()).toBe('Wi-Fi')
    expect(dataRows.at(4).text()).toBe('test')
    expect(dataRows.at(5).text()).toBe('test')
    expect(dataRows.at(6).text()).toBe('127.0.0.1')

    let buttons = wrapper.findAll('button')

    expect(buttons.length).toBe(2)

    expect(buttons.at(0).text()).toBe('Refresh Sessions')
    expect(buttons.at(1).text()).toBe('More Sessions')

    await buttons.at(0).trigger('click')

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(2)

    await localVue.nextTick()

    dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(7)

    expect(dataRows.at(0).text()).not.toBe('') // Formatted date based on test time
    expect(dataRows.at(1).text()).toBe('00000000')
    expect(dataRows.at(2).text()).toBe('test')
    expect(dataRows.at(3).text()).toBe('Wi-Fi')
    expect(dataRows.at(4).text()).toBe('test')
    expect(dataRows.at(5).text()).toBe('test')
    expect(dataRows.at(6).text()).toBe('127.0.0.1')

    expect(wrapper.vm.$data.currentPage).toBe(1)

    sessionsSpy = fetchUserSessionsMock(localVue, true, [
      {
        time_stamp: new Date(),
        meta: {
          id: '00000001',
          platform: 'test2',
          connection: 'wired',
          location: {
            isp: 'test2'
          },
          datacenter_alias: 'test2',
          server_addr: '127.0.0.2'
        }
      }
    ], MAX_USER_SESSION_PAGES, '00000000', 1)

    await buttons.at(1).trigger('click')

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(3)

    await localVue.nextTick()

    expect(wrapper.vm.$data.currentPage).toBe(MAX_USER_SESSION_PAGES)

    buttons = wrapper.findAll('button')

    expect(buttons.length).toBe(1)

    expect(buttons.at(0).text()).toBe('Refresh Sessions')

    dataRows = wrapper.findAll('tbody tr td')
    expect(dataRows.length).toBe(14)

    expect(dataRows.at(0).text()).not.toBe('') // Formatted date based on test time
    expect(dataRows.at(1).text()).toBe('00000000')
    expect(dataRows.at(2).text()).toBe('test')
    expect(dataRows.at(3).text()).toBe('Wi-Fi')
    expect(dataRows.at(4).text()).toBe('test')
    expect(dataRows.at(5).text()).toBe('test')
    expect(dataRows.at(6).text()).toBe('127.0.0.1')
    expect(dataRows.at(7).text()).not.toBe('') // Formatted date based on test time
    expect(dataRows.at(8).text()).toBe('00000001')
    expect(dataRows.at(9).text()).toBe('test2')
    expect(dataRows.at(10).text()).toBe('Wired')
    expect(dataRows.at(11).text()).toBe('test2')
    expect(dataRows.at(12).text()).toBe('test2')
    expect(dataRows.at(13).text()).toBe('127.0.0.2')

    sessionsSpy.mockReset()

    wrapper.destroy()
  })

  it('checks new looker functionality (new default behavior)', async () => {
    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)

    $route.path = '/user-tool/00000000'
    $route.params.pathMatch = '00000000'

    const totalSessions = 200

    let mockSessions = []

    for (let i = 0; i < totalSessions; i++) {
      // Add the mock session to the array
      mockSessions.push({
        timestamp: i, // Bad date but doesn't matter here
        meta: {
          id: `0000000${i}`,
          platform: 'test',
          connection: 'wifi',
          location: {
            isp: 'test'
          },
          datacenter_alias: 'test',
          server_addr: `127.0.0.${i}`
        }
      })
    }

    const sessionsSpy = fetchLookerUserSessionsMock(localVue, true, mockSessions, '00000000')

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const entriesDropDown = wrapper.find('#per-page-dropdown')
    expect(entriesDropDown.exists()).toBeTruthy()

    const paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    const skipBackwardButton = paginationNav.find('#skip-backward-button')
    expect(skipBackwardButton.exists()).toBeFalsy()

    const pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    const pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    const previousButton = paginationNav.find('#previous-page-button')
    expect(previousButton.exists()).toBeTruthy()
    expect(previousButton.classes('disabled')).toBeTruthy()

    const nextButton = paginationNav.find('#next-page-button')
    expect(nextButton.exists()).toBeTruthy()
    expect(nextButton.classes('disabled')).toBeFalsy()

    expect(pageLinks.at(1).text()).toBe('1') // First page after disabled previous button
    expect(pageItems.at(1).classes('active')).toBeTruthy()

    const skipForwardButton = paginationNav.find('#skip-forward-button')
    expect(skipForwardButton.exists()).toBeTruthy()

    expect(pageLinks.at(pageLinks.length - 3).text()).toBe('6') // 5 pages shown at a time - skip next and ... buttons

    const pageCounter = wrapper.find('#page-counter')
    expect(pageCounter.exists()).toBeTruthy()
    expect(pageCounter.text()).toBe('Total Pages: ' + Math.ceil(totalSessions / 10))

    const sessionTable = wrapper.find('table')
    expect(sessionTable.exists()).toBeTruthy()

    const tableHeaders = wrapper.findAll('th')
    expect(tableHeaders.length).toBe(7)

    expect(tableHeaders.at(0).text()).toBe('Date')
    expect(tableHeaders.at(1).text()).toBe('Session ID')
    expect(tableHeaders.at(2).text()).toBe('Platform')
    expect(tableHeaders.at(3).text()).toBe('Connection Type')
    expect(tableHeaders.at(4).text()).toBe('ISP')
    expect(tableHeaders.at(5).text()).toBe('Datacenter')
    expect(tableHeaders.at(6).text()).toBe('Server Address')

    const tableBody = sessionTable.find('tbody')
    expect(tableBody.exists()).toBeTruthy()

    const sessionEntries = tableBody.findAll('tr')
    expect(sessionEntries.length).toBe(10)

    sessionsSpy.mockReset()

    store.commit('UPDATE_IS_ADMIN', false)

    wrapper.destroy()
  })

  it('checks pagination behavior', async () => {
    jest.useFakeTimers()

    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)

    $route.path = '/user-tool/00000000'
    $route.params.pathMatch = '00000000'

    const totalSessions = 1995

    let mockSessions = []

    for (let i = 0; i < totalSessions; i++) {
      // Add the mock session to the array
      mockSessions.push({
        timestamp: i, // Bad date but doesn't matter here
        meta: {
          id: `0000000${i}`,
          platform: 'test',
          connection: 'wifi',
          location: {
            isp: 'test'
          },
          datacenter_alias: 'test',
          server_addr: `127.0.0.${i}`
        }
      })
    }

    const sessionsSpy = fetchLookerUserSessionsMock(localVue, true, mockSessions, '00000000')

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const entriesDropDown = wrapper.find('#per-page-dropdown')
    expect(entriesDropDown.exists()).toBeTruthy()

    let paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    let pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    let pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    let nextButton = paginationNav.find('#next-page-button')
    expect(nextButton.exists()).toBeTruthy()
    expect(nextButton.classes('disabled')).toBeFalsy()

    expect(pageLinks.at(1).text()).toBe('1') // First page after disabled previous button
    expect(pageItems.at(1).classes('active')).toBeTruthy()

    expect(pageLinks.at(pageLinks.length - 3).text()).toBe('6') // 5 pages shown at a time - skip next and ... buttons

    const sessionTable = wrapper.find('table')
    expect(sessionTable.exists()).toBeTruthy()

    let tableHeaders = wrapper.findAll('th')
    expect(tableHeaders.length).toBe(7)

    expect(tableHeaders.at(0).text()).toBe('Date')
    expect(tableHeaders.at(1).text()).toBe('Session ID')
    expect(tableHeaders.at(2).text()).toBe('Platform')
    expect(tableHeaders.at(3).text()).toBe('Connection Type')
    expect(tableHeaders.at(4).text()).toBe('ISP')
    expect(tableHeaders.at(5).text()).toBe('Datacenter')
    expect(tableHeaders.at(6).text()).toBe('Server Address')

    let tableBody = sessionTable.find('tbody')
    expect(tableBody.exists()).toBeTruthy()

    let sessionEntries = tableBody.findAll('tr')
    expect(sessionEntries.length).toBe(10)

    const secondPageLink = pageLinks.at(2)
    expect(secondPageLink.exists()).toBeTruthy()
    expect(secondPageLink.text()).toBe('2')

    await secondPageLink.trigger('click')

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    expect(pageLinks.at(1).text()).toBe('1') // First page after disabled previous button
    expect(pageLinks.at(2).text()).toBe('2') // Second page after disabled previous button
    expect(pageItems.at(2).classes('active')).toBeTruthy()

    expect(pageLinks.at(pageLinks.length - 3).text()).toBe('7') // 5 pages shown at a time - skip next and ... buttons

    tableBody = sessionTable.find('tbody')
    expect(tableBody.exists()).toBeTruthy()

    sessionEntries = tableBody.findAll('tr')
    expect(sessionEntries.length).toBe(10)

    const skipForwardButton = paginationNav.find('#skip-forward-button')
    expect(skipForwardButton.exists()).toBeTruthy()

    const skipForwardLink = skipForwardButton.find('a')
    expect(skipForwardLink.exists()).toBeTruthy()

    await skipForwardLink.trigger('click')

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    expect(pageLinks.at(Math.floor(pageLinks.length / 2)).text()).toBe('8') // middle of the stack of links
    expect(pageItems.at(Math.floor(pageLinks.length / 2)).classes('active')).toBeTruthy()

    expect(pageLinks.at(pageLinks.length - 3).text()).toBe('13') // 5 pages shown at a time - skip next and ... buttons

    tableBody = sessionTable.find('tbody')
    expect(tableBody.exists()).toBeTruthy()

    sessionEntries = tableBody.findAll('tr')
    expect(sessionEntries.length).toBe(10)

    nextButton = paginationNav.find('#next-page-button')
    expect(nextButton.exists()).toBeTruthy()
    expect(nextButton.classes('disabled')).toBeFalsy()

    let nextLink = nextButton.find('a')
    expect(nextLink.exists()).toBeTruthy()
    expect(nextLink.text()).toBe('Next')

    await nextLink.trigger('click')

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    expect(pageLinks.at(Math.floor(pageLinks.length / 2)).text()).toBe('9') // middle of the stack of links
    expect(pageItems.at(Math.floor(pageLinks.length / 2)).classes('active')).toBeTruthy()

    expect(pageLinks.at(pageLinks.length - 3).text()).toBe('14') // 5 pages shown at a time - skip next and ... buttons

    tableBody = sessionTable.find('tbody')
    expect(tableBody.exists()).toBeTruthy()

    sessionEntries = tableBody.findAll('tr')
    expect(sessionEntries.length).toBe(10)

    const previousButton = paginationNav.find('#previous-page-button')
    expect(nextButton.exists()).toBeTruthy()
    expect(nextButton.classes('disabled')).toBeFalsy()

    const previousLink = previousButton.find('a')
    expect(previousLink.exists()).toBeTruthy()
    expect(previousLink.text()).toBe('Previous')

    await previousLink.trigger('click')

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    expect(pageLinks.at(Math.floor(pageLinks.length / 2)).text()).toBe('8') // middle of the stack of links
    expect(pageItems.at(Math.floor(pageLinks.length / 2)).classes('active')).toBeTruthy()

    expect(pageLinks.at(pageLinks.length - 3).text()).toBe('13') // 5 pages shown at a time - skip next and ... buttons

    tableBody = sessionTable.find('tbody')
    expect(tableBody.exists()).toBeTruthy()

    sessionEntries = tableBody.findAll('tr')
    expect(sessionEntries.length).toBe(10)

    const skipBackwardButton = paginationNav.find('#skip-backward-button')
    expect(skipBackwardButton.exists()).toBeTruthy()

    const skipBackwardLink = skipBackwardButton.find('a')
    expect(skipBackwardLink.exists()).toBeTruthy()

    await skipBackwardLink.trigger('click')

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    expect(pageLinks.at(1).text()).toBe('1') // First page after disabled previous button
    expect(pageLinks.at(2).text()).toBe('2') // Second page after disabled previous button
    expect(pageItems.at(2).classes('active')).toBeTruthy()

    expect(pageLinks.at(pageLinks.length - 3).text()).toBe('7') // 5 pages shown at a time - skip next and ... buttons

    tableBody = sessionTable.find('tbody')
    expect(tableBody.exists()).toBeTruthy()

    sessionEntries = tableBody.findAll('tr')
    expect(sessionEntries.length).toBe(10)

    sessionsSpy.mockReset()

    store.commit('UPDATE_IS_ADMIN', false)

    wrapper.destroy()
  })

  it('checks entries per page drop down', async () => {
    jest.useFakeTimers()

    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)

    $route.path = '/user-tool/00000000'
    $route.params.pathMatch = '00000000'

    const totalSessions = 1995

    let mockSessions = []

    for (let i = 0; i < totalSessions; i++) {
      // Add the mock session to the array
      mockSessions.push({
        timestamp: i, // Bad date but doesn't matter here
        meta: {
          id: `0000000${i}`,
          platform: 'test',
          connection: 'wifi',
          location: {
            isp: 'test'
          },
          datacenter_alias: 'test',
          server_addr: `127.0.0.${i}`
        }
      })
    }

    const sessionsSpy = fetchLookerUserSessionsMock(localVue, true, mockSessions, '00000000')

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const entriesPerPage = wrapper.find('#per-page-dropdown')
    expect(entriesPerPage.exists()).toBeTruthy()

    const entries = entriesPerPage.findAll('option')
    expect(entries.length).toBe(5)

    let numPages = wrapper.find('#page-counter')
    expect(numPages.exists()).toBeTruthy()
    expect(numPages.text()).toBe(`Total Pages: ${Math.ceil(totalSessions / 10)}`)

    await entries.at(1).setSelected() // 25 sessions per page

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    numPages = wrapper.find('#page-counter')
    expect(numPages.exists()).toBeTruthy()
    expect(numPages.text()).toBe(`Total Pages: ${Math.ceil(totalSessions / 25)}`)

    await entries.at(entries.length - 1).setSelected() // 200 sessions per page

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    numPages = wrapper.find('#page-counter')
    expect(numPages.exists()).toBeTruthy()
    expect(numPages.text()).toBe(`Total Pages: ${Math.ceil(totalSessions / 200)}`)

    let paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    let pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    let pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    const skipForwardButton = paginationNav.find('#skip-forward-button')
    expect(skipForwardButton.exists()).toBeTruthy()

    const skipForwardLink = skipForwardButton.find('a')
    expect(skipForwardLink.exists()).toBeTruthy()

    await skipForwardLink.trigger('click')

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    const lastPageButton = pageLinks.at(pageLinks.length - 2) // Skip last button (next button)
    expect(lastPageButton.exists()).toBeTruthy()
    expect(lastPageButton.text()).toBe('10')

    await lastPageButton.trigger('click')

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    const sessionTable = wrapper.find('table')
    expect(sessionTable.exists()).toBeTruthy()

    const tableHeaders = wrapper.findAll('th')
    expect(tableHeaders.length).toBe(7)

    expect(tableHeaders.at(0).text()).toBe('Date')
    expect(tableHeaders.at(1).text()).toBe('Session ID')
    expect(tableHeaders.at(2).text()).toBe('Platform')
    expect(tableHeaders.at(3).text()).toBe('Connection Type')
    expect(tableHeaders.at(4).text()).toBe('ISP')
    expect(tableHeaders.at(5).text()).toBe('Datacenter')
    expect(tableHeaders.at(6).text()).toBe('Server Address')

    const tableBody = sessionTable.find('tbody')
    expect(tableBody.exists()).toBeTruthy()

    const sessionEntries = tableBody.findAll('tr')
    expect(sessionEntries.length).toBeLessThan(200) // 1995 isn't divisible by 200

    sessionsSpy.mockReset()

    store.commit('UPDATE_IS_ADMIN', false)

    wrapper.destroy()
  })

  it('checks per page counts', async () => {
    jest.useFakeTimers()

    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)

    $route.path = '/user-tool/00000000'
    $route.params.pathMatch = '00000000'

    const totalSessions = 1995

    let mockSessions = []

    for (let i = 0; i < totalSessions; i++) {
      // Add the mock session to the array
      mockSessions.push({
        timestamp: i, // Bad date but doesn't matter here
        meta: {
          id: `0000000${i}`,
          platform: 'test',
          connection: 'wifi',
          location: {
            isp: 'test'
          },
          datacenter_alias: 'test',
          server_addr: `127.0.0.${i}`
        }
      })
    }

    const sessionsSpy = fetchLookerUserSessionsMock(localVue, true, mockSessions, '00000000')

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const entriesPerPage = wrapper.find('#per-page-dropdown')
    expect(entriesPerPage.exists()).toBeTruthy()

    const entries = entriesPerPage.findAll('option')
    expect(entries.length).toBe(5)

    await entries.at(entries.length - 1).setSelected() // 200 sessions per page

    jest.advanceTimersByTime(2000)

    await localVue.nextTick()

    const numPages = wrapper.find('#page-counter')
    expect(numPages.exists()).toBeTruthy()
    expect(numPages.text()).toBe(`Total Pages: ${Math.ceil(totalSessions / 200)}`)

    let paginationNav = wrapper.find('.pagination')
    expect(paginationNav.exists()).toBeTruthy()

    const pageItems = paginationNav.findAll('li')
    expect(pageItems.length).toBeGreaterThan(0)

    let pageLinks = paginationNav.findAll('a')
    expect(pageLinks.length).toBeGreaterThan(0)

    const currentPageItem = pageItems.at(1) // skip next button
    expect(currentPageItem.exists()).toBeTruthy()
    expect(currentPageItem.classes('active')).toBeTruthy()

    const currentPageLink = pageLinks.at(1) // skip next button
    expect(currentPageLink.exists()).toBeTruthy()
    expect(currentPageLink.text()).toBe('1')

    for (let i = 2; i <= 10; i++) {
      const sessionTable = wrapper.find('table')
      expect(sessionTable.exists()).toBeTruthy()

      const tableBody = sessionTable.find('tbody')
      expect(tableBody.exists()).toBeTruthy()

      const sessionEntries = tableBody.findAll('tr')

      if (i < 10) {
        expect(sessionEntries.length).toBe(200)

        paginationNav = wrapper.find('.pagination')
        expect(paginationNav.exists()).toBeTruthy()

        pageLinks = paginationNav.findAll('a')
        expect(pageLinks.at(i).exists()).toBeTruthy()

        await pageLinks.at(i).trigger('click')

        jest.advanceTimersByTime(2000)

        await localVue.nextTick()
      } else {
        expect(sessionEntries.length).toBeLessThan(200)
        expect(sessionEntries.length).toBe(totalSessions % 200)
      }
    }

    sessionsSpy.mockReset()

    store.commit('UPDATE_IS_ADMIN', false)

    wrapper.destroy()
  })
})

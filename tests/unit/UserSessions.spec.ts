import {  createLocalVue, mount, shallowMount } from '@vue/test-utils'
import UserSessions from '@/components/UserSessions.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import Vuex from 'vuex'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { VueConstructor } from 'vue/types/umd'

function fetchUserSessionsMock (vueInstance: VueConstructor<any>, success: boolean, sessions: Array<any>, nextPage: number, userID: string, page: number): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchUserSessions').mockImplementation((args: any) => {
    expect(args.user_id).toBe(userID)
    expect(args.page).toBe(page)

    return success ? Promise.resolve({ sessions: sessions, page: nextPage }) : Promise.reject(new Error('Mock Error'))
  })
}

describe('UserSessions.vue no sessions', () => {
  const localVue = createLocalVue()

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

  const stubs = [
    'router-link'
  ]

  it('mounts the user sessions table successfully', () => {
    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks
    })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('checks default view with no sessions', async () => {
    $route.path = '/user-tool/00000000'
    $route.params.pathMatch = '00000000'

    const sessionsSpy = fetchUserSessionsMock(localVue, true, [], 10, '00000000', 0)

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks
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
    ], 10, '00000000', 0)

    const wrapper = shallowMount(UserSessions, {
      localVue, stubs, mocks
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
      localVue, stubs, mocks
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
    ], 10, '00000000', 1)

    await buttons.at(1).trigger('click')

    await localVue.nextTick()

    expect(sessionsSpy).toBeCalledTimes(3)

    await localVue.nextTick()

    expect(wrapper.vm.$data.currentPage).toBe(10)

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
})

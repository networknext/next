import { shallowMount, createLocalVue, Wrapper } from '@vue/test-utils'
import UserSessions from '@/components/UserSessions.vue'
import UserToolWorkspace from '@/workspaces/UserToolWorkspace.vue'
import VueRouter from 'vue-router'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { waitFor } from './utils'

describe('UserSessions.vue no sessions', () => {
  const localVue = createLocalVue()

  const defaultRouter = new VueRouter({
    routes: [
      {
        path: '/user-tool',
        name: 'user-tool',
        component: UserToolWorkspace,
        children: [
          {
            path: '*',
            name: 'user-sessions',
            component: UserSessions
          }
        ]
      }
    ]
  })

  localVue.use(VueRouter)
  localVue.use(JSONRPCPlugin)

  it('mounts the user sessions table successfully', () => {
    const router = defaultRouter
    const wrapper = shallowMount(UserSessions, { localVue, router })
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })

  it('check no sessions for user', async () => {
    const router = defaultRouter

    // Trigger the user sessions page through navigation
    router.push({ path: '/user-tool/0000000000' })

    // Check if navigation worked
    expect(router.currentRoute.fullPath).toBe('/user-tool/0000000000')

    const spy = jest.spyOn(localVue.prototype.$apiService, 'fetchUserSessions').mockImplementationOnce(() => {
      return Promise.resolve({ sessions: [] })
    })

    // Mount the component
    const wrapper = shallowMount(UserSessions, { localVue, router })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalled()

    // Search was successful and removed alert
    expect(wrapper.find('.alert').exists()).toBe(false)

    // Wait for the api call to come back and update the internal state
    await waitFor(wrapper, 'div')

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    // Grab all of the header column names and check if they are correct
    const headers = wrapper.findAll('th span')
    expect(headers.at(0).text()).toBe('Session ID')
    expect(headers.at(1).text()).toBe('Platform')
    expect(headers.at(2).text()).toBe('Connection Type')
    expect(headers.at(3).text()).toBe('ISP')
    expect(headers.at(4).text()).toBe('Datacenter')
    expect(headers.at(5).text()).toBe('Server Address')

    // Grab all of the data rows
    // Just 1 row -> shows no sessions
    const rows = wrapper.findAll('td')
    expect(wrapper.findAll('td').length).toBe(1)
    expect(rows.at(0).text()).toBe('There are no sessions belonging to this user.')
    wrapper.destroy()
  })

  it('check sessions for user', async () => {
    const router = new VueRouter({
      routes: [
        {
          path: '/user-tool',
          name: 'user-tool',
          component: UserToolWorkspace,
          children: [
            {
              path: '*',
              name: 'user-sessions',
              component: UserSessions
            }
          ]
        }
      ]
    })

    // Trigger the user sessions page through navigation
    router.push({ path: '/user-tool/0000000000' })

    // Check if navigation worked
    expect(router.currentRoute.fullPath).toBe('/user-tool/0000000000')

    const spy = jest.spyOn(localVue.prototype.$apiService, 'fetchUserSessions').mockImplementationOnce(() => {
      return Promise.resolve(
        {
          sessions: [{
            id: '123456789',
            platform: 'PS5',
            connection: 'star link',
            location: {
              isp: 'local'
            },
            datacenter: 'local',
            server_addr: '127.0.0.1'
          }]
        })
    })

    // Mount the component
    const wrapper = shallowMount(UserSessions, { localVue, router })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalled()

    // Search was successful and removed alert
    expect(wrapper.find('.alert').exists()).toBe(false)

    // Wait for the api call to come back and update the internal state
    await waitFor(wrapper, 'div')

    // Look for 1 table
    expect(wrapper.findAll('table').length).toBe(1)

    // Grab all of the header column names and check if they are correct
    const headers = wrapper.findAll('thead tr th span')
    expect(headers.at(0).text()).toBe('Session ID')
    expect(headers.at(1).text()).toBe('Platform')
    expect(headers.at(2).text()).toBe('Connection Type')
    expect(headers.at(3).text()).toBe('ISP')
    expect(headers.at(4).text()).toBe('Datacenter')
    expect(headers.at(5).text()).toBe('Server Address')

    // Grab all of the data rows
    // Just 1 row -> shows 1 session
    const rows = wrapper.findAll('#data-row')
    expect(rows.length).toBe(1)

    const row = wrapper.findAll('#data-row td')
    expect(row.at(0).text()).toBe('123456789')
    expect(row.at(1).text()).toBe('PS5')
    expect(row.at(2).text()).toBe('star link')
    expect(row.at(3).text()).toBe('local')
    expect(row.at(4).text()).toBe('local')
    expect(row.at(5).text()).toBe('127.0.0.1')
    wrapper.destroy()
  })
})

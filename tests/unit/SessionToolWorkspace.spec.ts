import { createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'

describe('SessionToolWorkspace.vue', () => {
  const localVue = createLocalVue()
  localVue.use(Vuex)

  const $route = {
    path: '/session-tool',
    params: {
      pathMatch: ''
    }
  }

  const mocks = {
    $route,
    $router: {
      push: function (newRoute: any) {
        $route.path = newRoute.path
      }
    }
  }

  const defaultStore = new Vuex.Store({
    state: {
      userProfile: {
        company: ''
      }
    },
    getters: {
      userProfile: (state: any) => state.userProfile,
      isAdmin: () => false,
      isOwner: () => false,
      registeredToCompany: () => false
    }
  })

  const stubs = [
    'router-view'
  ]

  it('mounts the user sessions table successfully', () => {
    const store = defaultStore
    const wrapper = mount(SessionToolWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })

  it('check default view', () => {
    // Mount the component
    const store = defaultStore
    const wrapper = mount(SessionToolWorkspace, {
      localVue, mocks, stubs, store
    })
    // Check Title
    expect(wrapper.find('.h2').text()).toBe('Session Tool')

    // Check label
    expect(wrapper.find('label').text()).toBe('Session ID')

    // Check for an input
    expect(wrapper.find('input').exists()).toBe(true)

    // Check input placeholder
    expect(wrapper.find('input').attributes('placeholder')).toBe('Enter a Session ID to view statistics')

    // Check button
    expect(wrapper.find('button').text()).toBe('View Stats')

    // Check for an info alert
    expect(wrapper.find('.alert').text()).toBe('Please enter a valid Session ID to view its statistics. It should be a hexadecimal number (with leading zeros), or a decimal number.')
    wrapper.destroy()
  })
})

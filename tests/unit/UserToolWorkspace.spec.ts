import { createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import UserToolWorkspace from '@/workspaces/UserToolWorkspace.vue'
import Alert from '@/components/Alert.vue'

describe('UserToolWorkspace.vue', () => {
  const localVue = createLocalVue()
  localVue.use(Vuex)

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
      isAnonymousPlus: () => false,
      registeredToCompany: () => false
    }
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
      push: function (newRoute: any) {
        $route.path = newRoute.path
      }
    }
  }

  const stubs = [
    'router-view'
  ]

  it('mounts the user sessions table successfully', () => {
    const store = defaultStore
    const wrapper = mount(UserToolWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })

  it('check no sessions for user', async () => {
    // Mount the component
    const store = defaultStore
    const wrapper = mount(UserToolWorkspace, {
      localVue, mocks, stubs, store
    })

    await localVue.nextTick()

    // Check Title
    expect(wrapper.find('.h2').text()).toBe('User Tool')

    // Check label
    expect(wrapper.find('label').text()).toBe('User ID')

    // Check for an input
    expect(wrapper.find('input').exists()).toBe(true)

    // Check input placeholder
    expect(wrapper.find('input').attributes('placeholder')).toBe('Enter a User ID to view their sessions')

    // Check button
    expect(wrapper.find('button').text()).toBe('View Sessions')

    // Check that the verify alert is hidden (no message)
    expect(wrapper.findAllComponents(Alert).at(0).text()).toBe('')
    // Check for an info alert
    expect(wrapper.findAllComponents(Alert).at(1).text()).toBe('Please enter a User ID to view their sessions.')
    wrapper.destroy()
  })
})

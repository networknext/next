import { createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import UserToolWorkspace from '@/workspaces/UserToolWorkspace.vue'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AlertType } from '@/components/types/AlertTypes'
import { ErrorTypes } from '@/components/types/ErrorTypes'

describe('UserToolWorkspace.vue', () => {
  const localVue = createLocalVue()
  localVue.use(Vuex)

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

  const defaultStore = {
    state: {
      userProfile: newDefaultProfile(),
      killLoops: false,
      isAnonymousPlus: false,
      isAdmin: false,
      isBuyer: false
    },
    getters: {
      allBuyers: (state: any) => state.allBuyers,
      userProfile: (state: any) => state.userProfile,
      isAnonymousPlus: (state: any) => state.isAnonymousPlus,
      killLoops: (state: any) => state.killLoops
    },
    actions: {
      toggleKillLoops ({ commit }: any, killLoops: boolean) {
        commit('TOGGLE_KILL_LOOPS', killLoops)
      }
    },
    mutations: {
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      },
      TOGGLE_KILL_LOOPS (state: any, killLoops: boolean) {
        state.killLoops = killLoops
      },
      UPDATE_IS_ANONYMOUSPLUS (state: any, isAnonymousPlus: boolean) {
        state.isAnonymousPlus = isAnonymousPlus
      }
    }
  }

  const stubs = [
    'router-view'
  ]

  it('mounts the user session tool successfully', () => {
    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(UserToolWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('check default view', async () => {
    // Mount the component
    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(UserToolWorkspace, {
      localVue, mocks, stubs, store
    })

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Please enter a User ID to view their sessions.')
    expect(alert.classes(AlertType.INFO)).toBeTruthy()

    // Check Title
    expect(wrapper.find('.h2').text()).toBe('User Tool')

    // Check label
    expect(wrapper.find('label').text()).toBe('User ID')

    // Check for an input
    expect(wrapper.find('input').exists()).toBeTruthy()

    // Check input placeholder
    expect(wrapper.find('input').attributes('placeholder')).toBe('Enter a User ID to view their sessions.')

    // Check button
    expect(wrapper.find('button').text()).toBe('View Sessions')

    wrapper.destroy()
  })

  it('checks kill loops alert', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('TOGGLE_KILL_LOOPS', true)

    const wrapper = mount(UserToolWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe(ErrorTypes.SYSTEM_FAILURE)

    store.commit('TOGGLE_KILL_LOOPS', false)

    wrapper.destroy()
  })

  it('checks invalid input', async () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = mount(UserToolWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    let alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Please enter a User ID to view their sessions.')
    expect(alert.classes(AlertType.INFO)).toBeTruthy()

    const idInput = wrapper.find('input')
    expect(idInput.exists()).toBeTruthy()

    await idInput.setValue('')

    const submitButton = wrapper.find('#user-tool-button')
    expect(submitButton.exists).toBeTruthy()

    await submitButton.trigger('submit')

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Please enter a User ID to view their sessions.')
    expect(alert.classes(AlertType.INFO)).toBeTruthy()

    await idInput.setValue('00000000')
    await submitButton.trigger('submit')

    expect($route.path).toBe('/user-tool/00000000')

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeFalsy()

    wrapper.destroy()
  })
})

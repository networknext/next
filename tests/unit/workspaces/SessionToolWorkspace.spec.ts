import { createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import SessionToolWorkspace from '@/workspaces/SessionToolWorkspace.vue'
import { EMAIL_CONFIRMATION_MESSAGE, SESSION_TOOL_ALERT } from '@/components/types/Constants'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AlertType } from '@/components/types/AlertTypes'
import { ErrorTypes } from '@/components/types/ErrorTypes'

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

  it('mounts the session tool successfully', () => {
    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(SessionToolWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('check default view', async () => {
    // Mount the component
    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(SessionToolWorkspace, {
      localVue, mocks, stubs, store
    })

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe(SESSION_TOOL_ALERT)
    expect(alert.classes(AlertType.INFO)).toBeTruthy()

    // Check Title
    expect(wrapper.find('.h2').text()).toBe('Session Tool')

    // Check label
    expect(wrapper.find('label').text()).toBe('Session ID')

    // Check for an input
    expect(wrapper.find('input').exists()).toBeTruthy()

    // Check input placeholder
    expect(wrapper.find('input').attributes('placeholder')).toBe('Enter a Session ID to view statistics')

    // Check button
    expect(wrapper.find('button').text()).toBe('View Stats')

    wrapper.destroy()
  })

  it('checks verification alert', async () => {
    const store = new Vuex.Store(defaultStore)

    const profile = newDefaultProfile()
    profile.email = 'test@test.com'

    store.commit('UPDATE_USER_PROFILE', profile)
    store.commit('UPDATE_IS_ANONYMOUSPLUS', true)

    const wrapper = mount(SessionToolWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.INFO)).toBeTruthy()
    expect(alert.text()).toContain(`${EMAIL_CONFIRMATION_MESSAGE} ${store.getters.userProfile.email}`)
    expect(alert.text()).toContain('Resend email')

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    store.commit('UPDATE_IS_ANONYMOUSPLUS', false)

    wrapper.destroy()
  })

  it('checks kill loops alert', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('TOGGLE_KILL_LOOPS', true)

    const wrapper = mount(SessionToolWorkspace, {
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

    const wrapper = mount(SessionToolWorkspace, {
      localVue, mocks, stubs, store
    })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    let alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe(SESSION_TOOL_ALERT)
    expect(alert.classes(AlertType.INFO)).toBeTruthy()

    const idInput = wrapper.find('input')
    expect(idInput.exists()).toBeTruthy()

    await idInput.setValue('')

    const submitButton = wrapper.find('#session-tool-button')
    expect(submitButton.exists).toBeTruthy()

    await submitButton.trigger('submit')

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe(SESSION_TOOL_ALERT)
    expect(alert.classes(AlertType.INFO)).toBeTruthy()

    await idInput.setValue('00000000')
    await submitButton.trigger('submit')

    expect($route.path).toBe('/session-tool/00000000')

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeFalsy()

    wrapper.destroy()
  })
})

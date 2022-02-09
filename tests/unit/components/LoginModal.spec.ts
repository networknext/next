import Vuex from 'vuex'
import { createLocalVue, shallowMount } from '@vue/test-utils'
import LoginModal from '@/components/LoginModal.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AuthPlugin } from '@/plugins/auth'
import { FeatureFlagService } from '@/plugins/flags'

describe('LoginModal.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultProfile = newDefaultProfile()

  const defaultStore = {
    state: {
      allBuyers: [],
      userProfile: defaultProfile
    },
    getters: {
      userProfile: (state: any) => state.userProfile
    },
    mutations: {
      UPDATE_ALL_BUYERS (state: any, allBuyers: Array<any>) {
        state.allBuyers = allBuyers
      },
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      }
    }
  }

  // Setup spy functions
  let windowSpy: jest.SpyInstance

  beforeEach(() => {
    windowSpy = jest.spyOn(window, 'window', 'get')
    windowSpy.mockImplementation(() => ({
      location: {
        origin: '127.0.0.1'
      }
    }))
  })

  afterEach(() => {
    windowSpy.mockRestore()
  })

  const $route = {
    path: '/login',
    params: {
      pathMatch: ''
    },
    query: {
      redirectURI: ''
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

  it('mounts a login modal successfully', () => {
    const wrapper = shallowMount(LoginModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()

    const header = wrapper.find('.header')
    expect(header.exists()).toBeTruthy()
    expect(header.text()).toBe('Log in')

    const inputs = wrapper.findAll('input')
    expect(inputs.length).toBe(2)

    const errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(0)

    const button = wrapper.find('button')
    expect(button.exists())
    expect(button.text()).toBe('Log in')

    const links = wrapper.findAll('router-link')
    expect(links.length).toBe(2)

    wrapper.destroy()
  })

  it('checks form layout', () => {
    const wrapper = shallowMount(LoginModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('checks form input validation', async () => {
    const store = new Vuex.Store(defaultStore)

    localVue.use(AuthPlugin, {
      domain: 'domain',
      clientID: 'clientID',
      store: store,
      flagService: new FeatureFlagService({
        flags: [],
        useAPI: false
      })
    })

    const login = jest.spyOn(localVue.prototype.$authService, 'login').mockImplementation(() => {
      return Promise.resolve()
    })

    const wrapper = shallowMount(LoginModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists).toBeTruthy()

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists).toBeTruthy()

    let errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(0)

    const button = wrapper.find('button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Log in')

    await button.trigger('submit')

    expect(login).not.toBeCalled()

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(1)
    expect(errors.at(0).text()).toBe('An email address is required')

    await emailInput.setValue('test@test.com')

    await button.trigger('submit')

    expect(login).not.toBeCalled()

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(1)
    expect(errors.at(0).text()).toBe('A password is required')

    login.mockReset()
    wrapper.destroy()
  })

  it('checks login failure', async () => {
    const store = new Vuex.Store(defaultStore)

    localVue.use(AuthPlugin, {
      domain: 'domain',
      clientID: 'clientID',
      store: store,
      flagService: new FeatureFlagService({
        flags: [],
        useAPI: false
      })
    })

    const login = jest.spyOn(localVue.prototype.$authService, 'login').mockImplementation(() => {
      return Promise.reject(new Error('Login Failed'))
    })

    const wrapper = shallowMount(LoginModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists).toBeTruthy()

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists).toBeTruthy()

    let errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(0)

    const button = wrapper.find('button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Log in')

    await emailInput.setValue('test@test.com')
    await passwordInput.setValue('superSecretPassword123')

    await button.trigger('submit')

    await localVue.nextTick()

    expect(login).toBeCalledTimes(1)

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(1)
    expect(errors.at(0).text()).toBe('Login Failed')

    login.mockReset()
    wrapper.destroy()
  })

  it('checks login success', async () => {
    const store = new Vuex.Store(defaultStore)

    localVue.use(AuthPlugin, {
      domain: 'domain',
      clientID: 'clientID',
      store: store,
      flagService: new FeatureFlagService({
        flags: [],
        useAPI: false
      })
    })

    const login = jest.spyOn(localVue.prototype.$authService, 'login').mockImplementation(() => {
      return Promise.resolve()
    })

    const wrapper = shallowMount(LoginModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists).toBeTruthy()

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists).toBeTruthy()

    let errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(0)

    const button = wrapper.find('button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Log in')

    await emailInput.setValue('test@test.com')
    await passwordInput.setValue('superSecretPassword123')

    await button.trigger('submit')

    expect(login).toBeCalledTimes(1)

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(0)

    login.mockReset()
    wrapper.destroy()
  })
})

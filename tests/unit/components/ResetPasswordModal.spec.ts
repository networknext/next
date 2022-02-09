import { createLocalVue, shallowMount } from '@vue/test-utils'
import ResetPasswordModal from '@/components/ResetPasswordModal.vue'
import { VueConstructor } from 'vue/types/umd'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

function sendResetPasswordEmailMock (
  vueInstance: VueConstructor<any>,
  success: boolean
): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'sendResetPasswordEmail').mockImplementation(() => {
    return success ? Promise.resolve() : Promise.reject(new Error('sendResetPasswordEmailMock Mock Error'))
  })
}

describe('ResetPasswordModal.vue', () => {
  const localVue = createLocalVue()

  localVue.use(JSONRPCPlugin)

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
    path: '/reset-password',
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

  it('mounts a password reset modal successfully', () => {
    const wrapper = shallowMount(ResetPasswordModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()

    const header = wrapper.find('.header')
    expect(header.exists()).toBeTruthy()
    expect(header.text()).toBe('Forgot Your Password?')

    const helperText = wrapper.find('p')
    expect(helperText.exists()).toBeTruthy()
    expect(helperText.text()).toBe('Enter your email address and we will send you instructions to reset your password.')

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBeTruthy()

    const continueButtons = wrapper.findAll('button')
    expect(continueButtons.length).toBe(1)

    const backToPortalLink = wrapper.find('router-link')
    expect(backToPortalLink.exists()).toBeTruthy()
    expect(backToPortalLink.text()).toBe('Back to Portal')

    wrapper.destroy()
  })

  it('checks form transition - success', async () => {
    const spy = sendResetPasswordEmailMock(localVue, true)

    const wrapper = shallowMount(ResetPasswordModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()

    expect(wrapper.vm.$data.stepOne).toBeTruthy()

    let header = wrapper.find('.header')
    expect(header.exists()).toBeTruthy()
    expect(header.text()).toBe('Forgot Your Password?')

    let helperText = wrapper.find('p')
    expect(helperText.exists()).toBeTruthy()
    expect(helperText.text()).toBe('Enter your email address and we will send you instructions to reset your password.')

    let emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBeTruthy()

    let continueButtons = wrapper.findAll('button')
    expect(continueButtons.length).toBe(1)

    let backToPortalLink = wrapper.find('router-link')
    expect(backToPortalLink.exists()).toBeTruthy()
    expect(backToPortalLink.text()).toBe('Back to Portal')

    await emailInput.setValue('test@test.com')

    await continueButtons.at(0).trigger('submit')

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    expect(wrapper.vm.$data.stepOne).toBeFalsy()

    header = wrapper.find('.header')
    expect(header.exists()).toBeTruthy()
    expect(header.text()).toBe('Check Your Email')

    helperText = wrapper.find('p')
    expect(helperText.exists()).toBeTruthy()
    expect(helperText.text()).toBe('Please check the email address ' + wrapper.vm.$data.email + ' for instructions to reset your password.')

    emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBeFalsy()

    continueButtons = wrapper.findAll('button')
    expect(continueButtons.length).toBe(1)
    expect(continueButtons.at(0).text()).toBe('Resend email')

    backToPortalLink = wrapper.find('router-link')
    expect(backToPortalLink.exists()).toBeTruthy()
    expect(backToPortalLink.text()).toBe('Back to Portal')

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks form transition - invalid email', async () => {
    const spy = sendResetPasswordEmailMock(localVue, true)

    const wrapper = shallowMount(ResetPasswordModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()

    expect(wrapper.vm.$data.stepOne).toBeTruthy()

    let header = wrapper.find('.header')
    expect(header.exists()).toBeTruthy()
    expect(header.text()).toBe('Forgot Your Password?')

    let helperText = wrapper.find('p')
    expect(helperText.exists()).toBeTruthy()
    expect(helperText.text()).toBe('Enter your email address and we will send you instructions to reset your password.')

    let emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBeTruthy()

    let continueButtons = wrapper.findAll('button')
    expect(continueButtons.length).toBe(1)

    let backToPortalLink = wrapper.find('router-link')
    expect(backToPortalLink.exists()).toBeTruthy()
    expect(backToPortalLink.text()).toBe('Back to Portal')

    await emailInput.setValue('test@')

    await continueButtons.at(0).trigger('submit')

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(0)

    const emailError = wrapper.find('small')
    expect(emailError.exists()).toBeTruthy()
    expect(emailError.text()).toBe('Please enter a valid email address')

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks form transition - failure', async () => {
    const spy = sendResetPasswordEmailMock(localVue, false)

    const wrapper = shallowMount(ResetPasswordModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()

    expect(wrapper.vm.$data.stepOne).toBeTruthy()

    let header = wrapper.find('.header')
    expect(header.exists()).toBeTruthy()
    expect(header.text()).toBe('Forgot Your Password?')

    let helperText = wrapper.find('p')
    expect(helperText.exists()).toBeTruthy()
    expect(helperText.text()).toBe('Enter your email address and we will send you instructions to reset your password.')

    let emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBeTruthy()

    let continueButtons = wrapper.findAll('button')
    expect(continueButtons.length).toBe(1)

    let backToPortalLink = wrapper.find('router-link')
    expect(backToPortalLink.exists()).toBeTruthy()
    expect(backToPortalLink.text()).toBe('Back to Portal')

    await emailInput.setValue('test@test.com')

    await continueButtons.at(0).trigger('submit')

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(1)

    const emailError = wrapper.find('small')
    expect(emailError.exists()).toBeTruthy()
    expect(emailError.text()).toBe('Could not send password reset email. Please verify that the email is linked to a valid account and try again')

    spy.mockReset()

    wrapper.destroy()
  })
})

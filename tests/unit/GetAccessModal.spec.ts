import Vuex from 'vuex'

import { createLocalVue, mount, shallowMount } from '@vue/test-utils'
import GetAccessModal from '@/components/GetAccessModal.vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faCheck } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { AuthPlugin } from '@/plugins/auth'
import { ErrorTypes } from '@/components/types/ErrorTypes'
import { FeatureFlagService, FlagPlugin } from '@/plugins/flags'
import { AlertType } from '@/components/types/AlertTypes'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

describe('GetAccessModal.vue', () => {
  const ICONS = [
    faCheck
  ]

  library.add(...ICONS)

  const localVue = createLocalVue()
  localVue.component('font-awesome-icon', FontAwesomeIcon)

  localVue.use(Vuex)

  const store = new Vuex.Store({
    state: {},
    getters: {},
    mutations: {}
  })

  localVue.use(AuthPlugin, {
    domain: 'domain',
    clientID: 'clientID',
    store: store,
    flagService: new FeatureFlagService({
      flags: [],
      useAPI: false
    })
  })

  localVue.use(JSONRPCPlugin)

  const $route = {
    query: ''
  }

  const mocks = {
    $route
  }

  const stubs = [
    'router-link'
  ]

  it('mounts a get access modal successfully', () => {
    const wrapper = shallowMount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('displays an error when an invalid email is submitted', async () => {
    const wrapper = shallowMount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBeTruthy()

    await emailInput.setValue('this is not a valid email')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    const inputElement = emailInput.element as HTMLInputElement
    expect(inputElement.value).toBe('this is not a valid email')

    await form.trigger('submit')

    let emailError = wrapper.find('#email-error')
    expect(emailError.text()).toBe(ErrorTypes.INVALID_EMAIL_ADDRESS)

    await emailInput.setValue('@badEmail.com')

    await form.trigger('submit')

    expect(emailError.text()).toBe(ErrorTypes.INVALID_EMAIL_ADDRESS)

    await emailInput.setValue('@.com')

    await form.trigger('submit')

    expect(emailError.text()).toBe(ErrorTypes.INVALID_EMAIL_ADDRESS)

    await emailInput.setValue('test@test.com')

    await form.trigger('submit')

    emailError = wrapper.find('#email-error')

    expect(emailError.exists()).toBeFalsy()

    wrapper.destroy()
  })

  it('password checker checks off the right check marks based on password', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists()).toBeTruthy()

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    await passwordInput.setValue('123')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes().style.includes('color: red;')).toBeTruthy()
    expect(wrapper.find('#lower-check').attributes().style.includes('color: red;')).toBeTruthy()
    expect(wrapper.find('#upper-check').attributes().style.includes('color: red;')).toBeTruthy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#special-check').attributes().style.includes('color: red;')).toBeTruthy()

    await passwordInput.setValue('12345678')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#length-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#lower-check').attributes().style.includes('color: red;')).toBeTruthy()
    expect(wrapper.find('#upper-check').attributes().style.includes('color: red;')).toBeTruthy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#special-check').attributes().style.includes('color: red;')).toBeTruthy()

    await passwordInput.setValue('a123456789')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#length-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#lower-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#lower-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#upper-check').attributes().style.includes('color: red;')).toBeTruthy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#special-check').attributes().style.includes('color: red;')).toBeTruthy()

    await passwordInput.setValue('abc123456789!')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#length-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#lower-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#lower-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#upper-check').attributes().style.includes('color: red;')).toBeTruthy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#special-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#special-check').attributes().style.includes('color: green;')).toBeTruthy()

    await passwordInput.setValue('Abc123456789!')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#length-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#lower-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#lower-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#upper-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#upper-check').attributes().style.includes('color: green;')).toBeTruthy()
    expect(wrapper.find('#number-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#special-check').attributes().style.includes('color: red;')).toBeFalsy()
    expect(wrapper.find('#special-check').attributes().style.includes('color: green;')).toBeTruthy()

    wrapper.destroy()
  })

  it('submits an invalid get access first page', async () => {
    const getAccessSpy = jest.spyOn(localVue.prototype.$authService, 'getAccess').mockImplementation(() => {
      return Promise.resolve()
    })

    const wrapper = shallowMount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBeTruthy()

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists()).toBeTruthy()

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    emailInput.setValue('@test.com')
    passwordInput.setValue('This is a bad password')

    await form.trigger('submit')

    expect(getAccessSpy).toBeCalledTimes(0)

    getAccessSpy.mockReset()
    wrapper.destroy()
  })

  it('submits a valid get access first page', async () => {
    const getAccessSpy = jest.spyOn(localVue.prototype.$authService, 'getAccess').mockImplementation(() => {
      return Promise.resolve()
    })

    const wrapper = shallowMount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBeTruthy()

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists()).toBeTruthy()

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    emailInput.setValue('test@test.com')
    passwordInput.setValue('Abcd1234567!?')

    await form.trigger('submit')

    expect(getAccessSpy).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    const secondPageHeader = wrapper.find('p')
    expect(secondPageHeader.text()).toBe('Please enter a company name and website so that our team can learn more about your company to help make your onboarding experience smoother.')

    getAccessSpy.mockReset()
    wrapper.destroy()
  })

  it('displays an error when an invalid first name is submitted', async () => {
    const wrapper = shallowMount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()

    // Skip the first page
    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const firstNameInput = wrapper.find('#first-name-input')
    expect(firstNameInput.exists()).toBeTruthy()

    firstNameInput.setValue('!@#!$@#')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    const inputElement = firstNameInput.element as HTMLInputElement
    expect(inputElement.value).toBe('!@#!$@#')

    await form.trigger('submit')

    let firstNameError = wrapper.find('#first-name-error')
    expect(firstNameError.text()).toBe(ErrorTypes.INVALID_FIRST_NAME)

    firstNameInput.setValue('this is a bad first name')

    await form.trigger('submit')

    firstNameError = wrapper.find('#first-name-error')
    expect(firstNameError.text()).toBe(ErrorTypes.INVALID_FIRST_NAME)

    firstNameInput.setValue('john')

    await form.trigger('submit')

    firstNameError = wrapper.find('#first-name-error')
    expect(firstNameError.exists()).toBeFalsy()

    wrapper.destroy()
  })

  it('displays an error when an invalid last name is submitted', async () => {
    const wrapper = shallowMount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const lastNameInput = wrapper.find('#last-name-input')
    expect(lastNameInput.exists()).toBeTruthy()

    lastNameInput.setValue('!@#!$@#')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    const inputElement = lastNameInput.element as HTMLInputElement
    expect(inputElement.value).toBe('!@#!$@#')

    await form.trigger('submit')

    let lastNameError = wrapper.find('#last-name-error')
    expect(lastNameError.text()).toBe(ErrorTypes.INVALID_LAST_NAME)

    lastNameInput.setValue('this is a bad last name')

    await form.trigger('submit')

    lastNameError = wrapper.find('#last-name-error')
    expect(lastNameError.text()).toBe(ErrorTypes.INVALID_LAST_NAME)

    lastNameInput.setValue('john')

    await form.trigger('submit')

    lastNameError = wrapper.find('#last-name-error')
    expect(lastNameError.exists()).toBeFalsy()

    wrapper.destroy()
  })

  it('displays an error when an invalid company name is submitted', async () => {
    const wrapper = shallowMount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const companyNameInput = wrapper.find('#company-name-input')
    expect(companyNameInput.exists()).toBeTruthy()

    companyNameInput.setValue('!@#!$@#')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    const inputElement = companyNameInput.element as HTMLInputElement
    expect(inputElement.value).toBe('!@#!$@#')

    await form.trigger('submit')

    let companyNameError = wrapper.find('#company-name-error')
    expect(companyNameError.text()).toBe(ErrorTypes.INVALID_COMPANY_NAME)

    companyNameInput.setValue('Test Company')

    await form.trigger('submit')

    companyNameError = wrapper.find('#company-name-error')
    expect(companyNameError.exists()).toBeFalsy()

    wrapper.destroy()
  })

  it('displays an error when an invalid company website is submitted', async () => {
    const wrapper = shallowMount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBeTruthy()

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const companyWebsiteInput = wrapper.find('#company-website-input')
    expect(companyWebsiteInput.exists()).toBeTruthy()

    companyWebsiteInput.setValue('!@#!$@#')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    const inputElement = companyWebsiteInput.element as HTMLInputElement
    expect(inputElement.value).toBe('!@#!$@#')

    await form.trigger('submit')

    let companyWebsiteError = wrapper.find('#company-website-error')
    expect(companyWebsiteError.text()).toBe(ErrorTypes.INVALID_WEBSITE)

    companyWebsiteInput.setValue('this is a bad website')

    await form.trigger('submit')

    companyWebsiteError = wrapper.find('#company-website-error')
    expect(companyWebsiteError.text()).toBe(ErrorTypes.INVALID_WEBSITE)

    companyWebsiteInput.setValue('https://networknext.com')

    await form.trigger('submit')

    companyWebsiteError = wrapper.find('#company-website-error')
    expect(companyWebsiteError.exists()).toBeFalsy()

    wrapper.destroy()
  })

  it('checks for alert on auto login failure', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    const processNewSignupSpy = jest.spyOn(localVue.prototype.$apiService, 'processNewSignup').mockImplementation(() => {
      return Promise.resolve()
    })

    const loginSpy = jest.spyOn(localVue.prototype.$authService, 'login').mockImplementation(() => {
      return Promise.reject()
    })

    expect(wrapper.exists()).toBeTruthy()

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const firstNameInput = wrapper.find('#first-name-input')
    expect(firstNameInput.exists()).toBeTruthy()

    firstNameInput.setValue('FirstName')

    const lastNameInput = wrapper.find('#last-name-input')
    expect(lastNameInput.exists()).toBeTruthy()

    lastNameInput.setValue('LastName')

    const companyNameInput = wrapper.find('#company-name-input')
    expect(companyNameInput.exists()).toBeTruthy()

    companyNameInput.setValue('Test Company')

    const companyWebsiteInput = wrapper.find('#company-website-input')
    expect(companyWebsiteInput.exists()).toBeTruthy()

    companyWebsiteInput.setValue('https://networknext.com')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    await form.trigger('submit')

    const firstNameError = wrapper.find('#first-name-error')
    expect(firstNameError.exists()).toBeFalsy()

    const lastNameError = wrapper.find('#last-name-error')
    expect(lastNameError.exists()).toBeFalsy()

    const companyNameError = wrapper.find('#company-name-error')
    expect(companyNameError.exists()).toBeFalsy()

    const companyWebsiteError = wrapper.find('#company-website-error')
    expect(companyWebsiteError.exists()).toBeFalsy()

    expect(processNewSignupSpy).toBeCalledTimes(1)
    expect(loginSpy).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    // Check for alert
    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe(ErrorTypes.AUTOMATIC_LOGIN_FAILURE + '\n  here to login manually')

    processNewSignupSpy.mockReset()
    loginSpy.mockReset()

    wrapper.destroy()
  })

  it('checks auto login success behavior', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    const processNewSignupSpy = jest.spyOn(localVue.prototype.$apiService, 'processNewSignup').mockImplementation(() => {
      return Promise.resolve()
    })

    const loginSpy = jest.spyOn(localVue.prototype.$authService, 'login').mockImplementation(() => {
      return Promise.resolve()
    })

    expect(wrapper.exists()).toBeTruthy()

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const firstNameInput = wrapper.find('#first-name-input')
    expect(firstNameInput.exists()).toBeTruthy()

    firstNameInput.setValue('FirstName')

    const lastNameInput = wrapper.find('#last-name-input')
    expect(lastNameInput.exists()).toBeTruthy()

    lastNameInput.setValue('LastName')

    const companyNameInput = wrapper.find('#company-name-input')
    expect(companyNameInput.exists()).toBeTruthy()

    companyNameInput.setValue('Test Company')

    const companyWebsiteInput = wrapper.find('#company-website-input')
    expect(companyWebsiteInput.exists()).toBeTruthy()

    companyWebsiteInput.setValue('https://networknext.com')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBeTruthy()

    await form.trigger('submit')

    const firstNameError = wrapper.find('#first-name-error')
    expect(firstNameError.exists()).toBeFalsy()

    const lastNameError = wrapper.find('#last-name-error')
    expect(lastNameError.exists()).toBeFalsy()

    const companyNameError = wrapper.find('#company-name-error')
    expect(companyNameError.exists()).toBeFalsy()

    const companyWebsiteError = wrapper.find('#company-website-error')
    expect(companyWebsiteError.exists()).toBeFalsy()

    expect(processNewSignupSpy).toBeCalledTimes(1)
    expect(loginSpy).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    // Check for alert
    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeFalsy()

    processNewSignupSpy.mockReset()
    loginSpy.mockReset()

    wrapper.destroy()

  })
})

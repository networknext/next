import { createLocalVue, mount } from '@vue/test-utils'
import GetAccessModal from '@/components/GetAccessModal.vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import { faCheck } from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { AuthPlugin } from '@/plugins/auth'
import { waitFor } from './utils'

describe('GetAccessModal.vue', () => {
  const ICONS = [
    faCheck
  ]

  library.add(...ICONS)

  const localVue = createLocalVue()
  localVue.component('font-awesome-icon', FontAwesomeIcon)

  localVue.use(AuthPlugin, {
    domain: 'domain',
    clientID: 'clientID'
  })

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
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })

  it('displays an error when an invalid email is submitted', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBe(true)

    await emailInput.setValue('this is not a valid email')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBe(true)

    const inputElement = <HTMLInputElement>emailInput.element
    expect(inputElement.value).toBe('this is not a valid email')

    await form.trigger('submit')

    let emailError = wrapper.find('#email-error')
    expect(emailError.text()).toBe('Please enter a valid email address')

    await emailInput.setValue('@badEmail.com')

    await form.trigger('submit')

    expect(emailError.text()).toBe('Please enter a valid email address')

    await emailInput.setValue('@.com')

    await form.trigger('submit')

    expect(emailError.text()).toBe('Please enter a valid email address')

    await emailInput.setValue('test@test.com')

    await form.trigger('submit')

    emailError = wrapper.find('#email-error')

    expect(emailError.exists()).toBe(false)

    wrapper.destroy()
  })

  it('password checker checks off the right check marks based on password', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists()).toBe(true)

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBe(true)

    await passwordInput.setValue('123')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes()['style'].includes('color: red;')).toBe(true)
    expect(wrapper.find('#lower-check').attributes()['style'].includes('color: red;')).toBe(true)
    expect(wrapper.find('#upper-check').attributes()['style'].includes('color: red;')).toBe(true)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#special-check').attributes()['style'].includes('color: red;')).toBe(true)

    await passwordInput.setValue('12345678')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#length-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#lower-check').attributes()['style'].includes('color: red;')).toBe(true)
    expect(wrapper.find('#upper-check').attributes()['style'].includes('color: red;')).toBe(true)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#special-check').attributes()['style'].includes('color: red;')).toBe(true)

    await passwordInput.setValue('a123456789')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#length-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#lower-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#lower-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#upper-check').attributes()['style'].includes('color: red;')).toBe(true)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#special-check').attributes()['style'].includes('color: red;')).toBe(true)

    await passwordInput.setValue('abc123456789!')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#length-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#lower-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#lower-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#upper-check').attributes()['style'].includes('color: red;')).toBe(true)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#special-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#special-check').attributes()['style'].includes('color: green;')).toBe(true)

    await passwordInput.setValue('Abc123456789!')

    await form.trigger('submit')

    expect(wrapper.find('#length-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#length-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#lower-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#lower-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#upper-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#upper-check').attributes()['style'].includes('color: green;')).toBe(true)
    expect(wrapper.find('#number-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#special-check').attributes()['style'].includes('color: red;')).toBe(false)
    expect(wrapper.find('#special-check').attributes()['style'].includes('color: green;')).toBe(true)

    wrapper.destroy()
  })

  it('submits an invalid get access first page', async () => {
    const getAccessSpy = jest.spyOn(localVue.prototype.$authService, 'getAccess').mockImplementation(() => {
      return Promise.resolve()
    })

    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBe(true)

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists()).toBe(true)

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBe(true)

    emailInput.setValue('@test.com')
    passwordInput.setValue('This is a bad password')

    await form.trigger('submit')

    expect(getAccessSpy).toBeCalledTimes(0)
  })

  it('submits a valid get access first page', async () => {
    const getAccessSpy = jest.spyOn(localVue.prototype.$authService, 'getAccess').mockImplementation(() => {
      return Promise.resolve()
    })

    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)

    const emailInput = wrapper.find('#email-input')
    expect(emailInput.exists()).toBe(true)

    const passwordInput = wrapper.find('#password-input')
    expect(passwordInput.exists()).toBe(true)

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBe(true)

    emailInput.setValue('test@test.com')
    passwordInput.setValue('Abcd1234567!?')

    await form.trigger('submit')

    expect(getAccessSpy).toBeCalledTimes(1)

    await wrapper.vm.$nextTick()

    const secondPageHeader = wrapper.find('p')
    expect(secondPageHeader.text()).toBe('Please enter a company name and website so that our team can learn more about your company to help make your on boarding experience smoother.')
  })

  it('displays an error when an invalid first name is submitted', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const firstNameInput = wrapper.find('#first-name-input')
    expect(firstNameInput.exists()).toBe(true)

    firstNameInput.setValue('!@#!$@#')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBe(true)

    const inputElement = <HTMLInputElement>firstNameInput.element
    expect(inputElement.value).toBe('!@#!$@#')

    await form.trigger('submit')

    let firstNameError = wrapper.find('#first-name-error')
    expect(firstNameError.text()).toBe('Please enter a valid first name')

    firstNameInput.setValue('this is a bad first name')

    await form.trigger('submit')

    firstNameError = wrapper.find('#first-name-error')
    expect(firstNameError.text()).toBe('Please enter a valid first name')

    firstNameInput.setValue('john')

    await form.trigger('submit')

    firstNameError = wrapper.find('#first-name-error')
    expect(firstNameError.exists()).toBe(false)
  })

  it('displays an error when an invalid last name is submitted', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const lastNameInput = wrapper.find('#last-name-input')
    expect(lastNameInput.exists()).toBe(true)

    lastNameInput.setValue('!@#!$@#')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBe(true)

    const inputElement = <HTMLInputElement>lastNameInput.element
    expect(inputElement.value).toBe('!@#!$@#')

    await form.trigger('submit')

    let lastNameError = wrapper.find('#last-name-error')
    expect(lastNameError.text()).toBe('Please enter a valid last name')

    lastNameInput.setValue('this is a bad last name')

    await form.trigger('submit')

    lastNameError = wrapper.find('#last-name-error')
    expect(lastNameError.text()).toBe('Please enter a valid last name')

    lastNameInput.setValue('john')

    await form.trigger('submit')

    lastNameError = wrapper.find('#last-name-error')
    expect(lastNameError.exists()).toBe(false)
  })

  it('displays an error when an invalid company name is submitted', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const companyNameInput = wrapper.find('#company-name-input')
    expect(companyNameInput.exists()).toBe(true)

    companyNameInput.setValue('!@#!$@#')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBe(true)

    const inputElement = <HTMLInputElement>companyNameInput.element
    expect(inputElement.value).toBe('!@#!$@#')

    await form.trigger('submit')

    let companyNameError = wrapper.find('#company-name-error')
    expect(companyNameError.text()).toBe('Please enter a valid company name')

    companyNameInput.setValue('Test Company')

    await form.trigger('submit')

    companyNameError = wrapper.find('#company-name-error')
    expect(companyNameError.exists()).toBe(false)
  })

  it('displays an error when an invalid company website is submitted', async () => {
    const wrapper = mount(GetAccessModal, {
      localVue, mocks, stubs
    })

    expect(wrapper.exists()).toBe(true)

    wrapper.vm.$data.stepOne = false

    await wrapper.vm.$nextTick()

    const companyWebsiteInput = wrapper.find('#company-website-input')
    expect(companyWebsiteInput.exists()).toBe(true)

    companyWebsiteInput.setValue('!@#!$@#')

    const form = wrapper.find('#get-access-form')
    expect(form.exists()).toBe(true)

    const inputElement = <HTMLInputElement>companyWebsiteInput.element
    expect(inputElement.value).toBe('!@#!$@#')

    await form.trigger('submit')

    let companyWebsiteError = wrapper.find('#company-website-error')
    expect(companyWebsiteError.text()).toBe('Please enter a valid website. IE: https://networknext.com')

    companyWebsiteInput.setValue('this is a bad website')

    await form.trigger('submit')

    companyWebsiteError = wrapper.find('#company-website-error')
    expect(companyWebsiteError.text()).toBe('Please enter a valid website. IE: https://networknext.com')

    companyWebsiteInput.setValue('https://networknext.com')

    await form.trigger('submit')

    companyWebsiteError = wrapper.find('#company-website-error')
    expect(companyWebsiteError.exists()).toBe(false)
  })
})

import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import AccountSettings from '@/components/AccountSettings.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AuthPlugin } from '@/plugins/auth'
import { FeatureFlagService } from '@/plugins/flags'
import { AlertType } from '@/components/types/AlertTypes'

describe('AccountSettings.vue', () => {
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
      allBuyers: (state: any) => state.allBuyers,
      isOwner: () => (state: any) => state.userProfile.roles.indexOf('Owner') !== -1,
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

  // Run bare minimum mount test
  it('mounts the component successfully', () => {
    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('checks general form', () => {
    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const titles = wrapper.findAll('.card-title')
    expect(titles.length).toBe(2)
    expect(titles.at(0).text()).toBe('User Details')
    expect(titles.at(1).text()).toBe('Company Details')

    const details = wrapper.findAll('.card-text')
    expect(details.length).toBe(2)
    expect(details.at(0).text()).toBe('Update user account profile.')
    expect(details.at(1).text()).toBe('Create or assign yourself to a company account.')

    const labels = wrapper.findAll('label')
    expect(labels.length).toBe(4)
    expect(labels.at(0).text()).toBe('First Name')
    expect(labels.at(1).text()).toBe('Last Name')
    expect(labels.at(2).text()).toBe('Company Name')
    expect(labels.at(3).text()).toBe('Company Code')

    const inputs = wrapper.findAll('input')
    expect(inputs.length).toBe(5)

    const inputDetails = wrapper.findAll('small')
    expect(inputDetails.length).toBe(3)
    expect(inputDetails.at(0).text()).toBe('I would like to receive the Network Next newsletter')
    expect(inputDetails.at(1).text()).toBe('This is the name of the company that you would like your account to be assigned to. This is not necessary for existing company assignment and is case and white space sensitive.')
    expect(inputDetails.at(2).text()).toBe('This is the unique string associated to your company account and to be used in your company\'s subdomain. To assign this user account to an existing company, type in your companies existing code. Examples: mycompany, my-company, my-company-name')

    const buttons = wrapper.findAll('button')
    expect(buttons.length).toBe(2)
    expect(buttons.at(0).text()).toBe('Update User Details')
    expect(buttons.at(1).text()).toBe('Setup Company Account')
    wrapper.destroy()
  })

  it('checks valid input values - user account details', async () => {
    const updateAccountDetails = jest.spyOn(localVue.prototype.$apiService, 'updateAccountDetails').mockImplementation(() => {
      return Promise.resolve({})
    })

    const setupCompanyAccount = jest.spyOn(localVue.prototype.$apiService, 'setupCompanyAccount').mockImplementation(() => {
      return Promise.resolve({})
    })

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

    const refreshToken = jest.spyOn(localVue.prototype.$authService, 'refreshToken').mockImplementation(() => {
      return Promise.resolve()
    })

    const tooLongString = '!'.padStart(2049, '!')
    expect(tooLongString.length).toBe(2049)

    const wrapper = mount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const firstNameInput = wrapper.find('#firstNameInput')
    expect(firstNameInput.exists()).toBeTruthy()

    const lastNameInput = wrapper.find('#lastNameInput')
    expect(lastNameInput.exists()).toBeTruthy()

    const button = wrapper.find('#account-details-button')
    expect(button.text()).toBe('Update User Details')

    await button.trigger('submit')

    let errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(2)
    expect(errors.at(0).text()).toBe('Please enter your first name')
    expect(errors.at(1).text()).toBe('Please enter your last name')
    expect(updateAccountDetails).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await firstNameInput.setValue('FirstName')

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(1)
    expect(errors.at(0).text()).toBe('Please enter your last name')
    expect(updateAccountDetails).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await firstNameInput.setValue('')
    await lastNameInput.setValue('LastName')

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(1)
    expect(errors.at(0).text()).toBe('Please enter your first name')
    expect(updateAccountDetails).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await firstNameInput.setValue('!!!!!!!!!!!!!!!!!!')

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(1)
    expect(errors.at(0).text()).toBe('A valid first name must include at least one letter')
    expect(updateAccountDetails).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await firstNameInput.setValue(tooLongString)

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(1)
    expect(errors.at(0).text()).toBe('First name is to long, please enter a name that is less that 2048 characters')
    expect(updateAccountDetails).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await firstNameInput.setValue('')
    await lastNameInput.setValue('!!!!!!!!!!!!!!!!!!')

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(2)
    expect(errors.at(0).text()).toBe('Please enter your first name')
    expect(errors.at(1).text()).toBe('A valid last name must include at least one letter')
    expect(updateAccountDetails).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await lastNameInput.setValue(tooLongString)

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(2)
    expect(errors.at(0).text()).toBe('Please enter your first name')
    expect(errors.at(1).text()).toBe('Last name is to long, please enter a name that is less that 2048 characters')
    expect(updateAccountDetails).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await firstNameInput.setValue('FirstName')
    await lastNameInput.setValue('LastName')

    await button.trigger('submit')

    await localVue.nextTick()
    await localVue.nextTick()

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(0)
    expect(updateAccountDetails).toBeCalledTimes(1)
    expect(refreshToken).toBeCalledTimes(1)

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Account details updated successfully')
    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()

    refreshToken.mockReset()
    updateAccountDetails.mockReset()
    setupCompanyAccount.mockReset()
    wrapper.destroy()
  })

  it('checks valid input values - company details', async () => {
    const updateAccountDetails = jest.spyOn(localVue.prototype.$apiService, 'updateAccountDetails').mockImplementation(() => {
      return Promise.resolve({})
    })

    const setupCompanyAccount = jest.spyOn(localVue.prototype.$apiService, 'setupCompanyAccount').mockImplementation(() => {
      return Promise.resolve({})
    })

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

    const refreshToken = jest.spyOn(localVue.prototype.$authService, 'refreshToken').mockImplementation(() => {
      return Promise.resolve()
    })

    const tooLongString = '!'.padStart(2048, '!')
    expect(tooLongString.length).toBe(2048)

    const wrapper = mount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const companyNameInput = wrapper.find('#companyNameInput')
    expect(companyNameInput.exists()).toBeTruthy()

    const companyCodeInput = wrapper.find('#companyCodeInput')
    expect(companyCodeInput.exists()).toBeTruthy()

    const button = wrapper.find('#company-details-button')
    expect(button.text()).toBe('Setup Company Account')

    await button.trigger('submit')

    let errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(2)
    expect(errors.at(0).text()).toBe('If setting up a company account, please enter a company name and code, otherwise please enter a valid company code')
    expect(errors.at(1).text()).toBe('Please choose a company code that contains character padded hyphens and no special characters')
    expect(setupCompanyAccount).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await companyNameInput.setValue('Company Name')

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(2)
    expect(errors.at(0).text()).toBe('A company code is required for company creation / assignment')
    expect(setupCompanyAccount).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await companyNameInput.setValue(tooLongString)

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(3)
    expect(errors.at(0).text()).toBe('Please choose a company name that is at most 256 characters')
    expect(errors.at(1).text()).toBe('A company code is required for company creation / assignment')
    expect(errors.at(2).text()).toBe('Please choose a company code that contains character padded hyphens and no special characters')
    expect(setupCompanyAccount).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await companyCodeInput.setValue('Bad company code input')

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(2)
    expect(errors.at(0).text()).toBe('Please choose a company name that is at most 256 characters')
    expect(errors.at(1).text()).toBe('Please choose a company code that contains character padded hyphens and no special characters')
    expect(setupCompanyAccount).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await companyCodeInput.setValue('!'.padStart(33, '!'))

    await button.trigger('submit')

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(3)
    expect(errors.at(0).text()).toBe('Please choose a company name that is at most 256 characters')
    expect(errors.at(1).text()).toBe('Please choose a company code that is at most 32 characters')
    expect(errors.at(2).text()).toBe('Please choose a company code that contains character padded hyphens and no special characters')
    expect(setupCompanyAccount).not.toBeCalled()
    expect(refreshToken).not.toBeCalled()

    await companyNameInput.setValue('Test Company')
    await companyCodeInput.setValue('test')

    await button.trigger('submit')

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Please update your first and last name before setting up a company account')
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()

    errors = wrapper.findAll('.text-danger')
    expect(errors.length).toBe(0)
    expect(setupCompanyAccount).toBeCalledTimes(0)
    expect(refreshToken).toBeCalledTimes(0)

    refreshToken.mockReset()
    updateAccountDetails.mockReset()
    setupCompanyAccount.mockReset()
    wrapper.destroy()
  })

  it('checks user details update - failure', async () => {
    const updateAccountDetails = jest.spyOn(localVue.prototype.$apiService, 'updateAccountDetails').mockImplementation(() => {
      return Promise.reject(new Error('updateAccountDetails error'))
    })

    const refreshToken = jest.spyOn(localVue.prototype.$authService, 'refreshToken').mockImplementation(() => {
      return Promise.reject(new Error('updateAccountDetails error'))
    })

    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const firstNameInput = wrapper.find('#firstNameInput')
    expect(firstNameInput.exists()).toBeTruthy()

    const lastNameInput = wrapper.find('#lastNameInput')
    expect(lastNameInput.exists()).toBeTruthy()

    const button = wrapper.find('#account-details-button')
    expect(button.text()).toBe('Update User Details')

    await firstNameInput.setValue('FirstName')
    await lastNameInput.setValue('LastName')

    await button.trigger('submit')

    await localVue.nextTick()

    expect(updateAccountDetails).toBeCalledTimes(1)
    expect(refreshToken).not.toBeCalled()

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Failed to update account details')
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()

    updateAccountDetails.mockReset()
    refreshToken.mockReset()
    wrapper.destroy()
  })

  it('checks user details update - success', async () => {
    const updateAccountDetails = jest.spyOn(localVue.prototype.$apiService, 'updateAccountDetails').mockImplementation(() => {
      return Promise.resolve()
    })

    const refreshToken = jest.spyOn(localVue.prototype.$authService, 'refreshToken').mockImplementation(() => {
      return Promise.resolve()
    })

    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const firstNameInput = wrapper.find('#firstNameInput')
    expect(firstNameInput.exists()).toBeTruthy()

    const lastNameInput = wrapper.find('#lastNameInput')
    expect(lastNameInput.exists()).toBeTruthy()

    const button = wrapper.find('#account-details-button')
    expect(button.text()).toBe('Update User Details')

    await firstNameInput.setValue('FirstName')
    await lastNameInput.setValue('LastName')

    await button.trigger('submit')

    await localVue.nextTick()

    expect(updateAccountDetails).toBeCalledTimes(1)
    expect(refreshToken).toBeCalledTimes(1)

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Account details updated successfully')
    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()

    updateAccountDetails.mockReset()
    refreshToken.mockReset()
    wrapper.destroy()
  })

  it('checks company account creation - failure', async () => {
    const setupCompanyAccount = jest.spyOn(localVue.prototype.$apiService, 'setupCompanyAccount').mockImplementation(() => {
      return Promise.reject(new Error('updateAccountDetails error'))
    })

    const refreshToken = jest.spyOn(localVue.prototype.$authService, 'refreshToken').mockImplementation(() => {
      return Promise.reject(new Error('updateAccountDetails error'))
    })

    const store = new Vuex.Store(defaultStore)

    const profile = newDefaultProfile()

    profile.firstName = 'FirstName'
    profile.lastName = 'LastName'
    store.commit('UPDATE_USER_PROFILE', profile)

    const wrapper = mount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const companyNameInput = wrapper.find('#companyNameInput')
    expect(companyNameInput.exists()).toBeTruthy()

    const companyCodeInput = wrapper.find('#companyCodeInput')
    expect(companyCodeInput.exists()).toBeTruthy()

    const button = wrapper.find('#company-details-button')
    expect(button.text()).toBe('Setup Company Account')

    await companyNameInput.setValue('Test Company')
    await companyCodeInput.setValue('test')

    await button.trigger('submit')

    await localVue.nextTick()

    expect(setupCompanyAccount).toBeCalledTimes(1)
    expect(refreshToken).not.toBeCalled()

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Failed to update company details')
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())

    setupCompanyAccount.mockReset()
    refreshToken.mockReset()
    wrapper.destroy()
  })

  it('checks company account creation - success', async () => {
    const setupCompanyAccount = jest.spyOn(localVue.prototype.$apiService, 'setupCompanyAccount').mockImplementation(() => {
      return Promise.resolve()
    })

    const refreshToken = jest.spyOn(localVue.prototype.$authService, 'refreshToken').mockImplementation(() => {
      return Promise.resolve()
    })

    const store = new Vuex.Store(defaultStore)

    const profile = newDefaultProfile()

    profile.firstName = 'FirstName'
    profile.lastName = 'LastName'
    store.commit('UPDATE_USER_PROFILE', profile)

    const wrapper = mount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const companyNameInput = wrapper.find('#companyNameInput')
    expect(companyNameInput.exists()).toBeTruthy()

    const companyCodeInput = wrapper.find('#companyCodeInput')
    expect(companyCodeInput.exists()).toBeTruthy()

    const button = wrapper.find('#company-details-button')
    expect(button.text()).toBe('Setup Company Account')

    await companyNameInput.setValue('Test Company')
    await companyCodeInput.setValue('test')

    await button.trigger('submit')

    await localVue.nextTick()

    expect(setupCompanyAccount).toBeCalledTimes(1)
    expect(refreshToken).toBeCalledTimes(1)

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.text()).toBe('Account settings updated successfully')
    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())

    setupCompanyAccount.mockReset()
    refreshToken.mockReset()
    wrapper.destroy()
  })
})

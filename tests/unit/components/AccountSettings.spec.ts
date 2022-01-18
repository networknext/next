import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import AccountSettings from '@/components/AccountSettings.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'

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

  it('checks user details update - failure', () => {
    const updateAccountDetails = jest.spyOn(localVue.prototype.$apiService, 'updateAccountDetails').mockImplementationOnce(() => {
      return Promise.reject(new Error('updateAccountDetails error'))
    })

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    updateAccountDetails.mockReset()
    wrapper.destroy()
  })

  it('checks valid input values', () => {
    const updateAccountDetails = jest.spyOn(localVue.prototype.$apiService, 'updateAccountDetails').mockImplementationOnce(() => {
      return Promise.resolve({})
    })

    const setupCompanyAccount = jest.spyOn(localVue.prototype.$apiService, 'setupCompanyAccount').mockImplementationOnce(() => {
      return Promise.resolve({})
    })

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    updateAccountDetails.mockReset()
    setupCompanyAccount.mockReset()
    wrapper.destroy()
  })

  it('checks user details update - success', () => {
    const updateAccountDetails = jest.spyOn(localVue.prototype.$apiService, 'updateAccountDetails').mockImplementationOnce(() => {
      return Promise.resolve({})
    })

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    updateAccountDetails.mockReset()
    wrapper.destroy()
  })

  it('checks company account creation - failure', () => {
    const setupCompanyAccount = jest.spyOn(localVue.prototype.$apiService, 'setupCompanyAccount').mockImplementationOnce(() => {
      return Promise.reject(new Error('setupCompanyAccount error'))
    })

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    setupCompanyAccount.mockReset()
    wrapper.destroy()
  })

  it('checks company account creation - success', () => {
    const setupCompanyAccount = jest.spyOn(localVue.prototype.$apiService, 'setupCompanyAccount').mockImplementationOnce(() => {
      return Promise.resolve({})
    })

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(AccountSettings, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    setupCompanyAccount.mockReset()
    wrapper.destroy()
  })
})

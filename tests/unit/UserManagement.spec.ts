import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import UserManagement from '@/components/UserManagement.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { VueConstructor } from 'vue/types/umd'
import { AlertType } from '@/components/types/AlertTypes'

function allAccountsMock (vueInstance: VueConstructor<any>, success: boolean, accounts: Array<any>): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchAllAccounts').mockImplementation(() => {
    return success ? Promise.resolve(
      {
        accounts: accounts
      }
    ) : Promise.reject(new Error('Mock Error'))
  })
}

function allRolesMock (vueInstance: VueConstructor<any>, success: boolean, roles: Array<any>): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchAllRoles').mockImplementation(() => {
    return success ? Promise.resolve(
      {
        roles: roles
      }
    ) : Promise.reject(new Error('Mock Error'))
  })
}

function updateAutoSignupDomainsMock (vueInstance: VueConstructor<any>, success: boolean, domains: Array<string>): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'updateAutoSignupDomains').mockImplementation(() => {
    return success ? Promise.resolve(
      {
        domains: domains
      }
    ) : Promise.reject(new Error('Mock Error'))
  })
}

describe('UserManagement.vue', () => {
  const localVue = createLocalVue()

  // Setup plugins
  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  // Init the store instance
  const defaultStore = {
    state: {
      userProfile: newDefaultProfile()
    },
    getters: {
      userProfile: (state: any) => state.userProfile,
      isAdmin: (state: any) => state.userProfile.roles.indexOf('Admin') !== -1,
      isOwner: (state: any, getters: any) => state.userProfile.roles.indexOf('Owner') !== -1 || getters.isAdmin,
      registeredToCompany: (state: any) => (state.userProfile.companyCode !== '')
    },
    mutations: {
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      },
      UPDATE_USER_ROLES (state: any, roles: Array<string>) {
        state.userProfile.roles = roles
      },
      UPDATE_USER_CUSTOMER_CODE (state: any, customerCode: string) {
        state.userProfile.companyCode = customerCode
      }
    }
  }

  it('mounts the component successfully', () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()

    wrapper.destroy()
  })

  // TODO: The tests below and their associated forms can be moved to their own components eventually
  it('checks auto sign up form - empty', async () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)

    const newProfile = newDefaultProfile()
    newProfile.companyCode = 'test'
    store.commit('UPDATE_USER_PROFILE', newProfile)

    const wrapper = shallowMount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const form = wrapper.find('#auto-signup-form')
    expect(form.exists()).toBeTruthy()

    const label = form.find('label')
    expect(label.exists()).toBeTruthy()
    expect(label.text()).toBe('Automatic Sign up Domains')

    const helperText = form.find('small')
    expect(helperText.exists()).toBeTruthy()
    expect(helperText.text()).toBe('Setting this to a comma seperated list of email domains will allow anyone with that domain to assign themselves to your account using your company code (test) in the account settings page.')

    const textArea = form.find('#auto-signup-domains')
    const textAreaElement = textArea.element as HTMLTextAreaElement
    expect(textArea.exists()).toBeTruthy()
    expect(textAreaElement.value).toBe('')

    const button = form.find('#auto-signup-button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Save Automatic Sign up')

    wrapper.destroy()

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks auto sign up form - !empty', async () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)

    const newProfile = newDefaultProfile()
    newProfile.domains = ['test.com']
    store.commit('UPDATE_USER_PROFILE', newProfile)

    const wrapper = shallowMount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const form = wrapper.find('#auto-signup-form')
    expect(form.exists()).toBeTruthy()

    const label = form.find('label')
    expect(label.exists()).toBeTruthy()
    expect(label.text()).toBe('Automatic Sign up Domains')

    const textArea = form.find('#auto-signup-domains')
    const textAreaElement = textArea.element as HTMLTextAreaElement
    expect(textArea.exists()).toBeTruthy()
    expect(textAreaElement.value).toBe('test.com')

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks auto sign up form - !empty - multiple', async () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)
    const newProfile = newDefaultProfile()
    newProfile.domains = ['test.com', 'google.com', 'networknext.com']
    store.commit('UPDATE_USER_PROFILE', newProfile)

    const wrapper = shallowMount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const form = wrapper.find('#auto-signup-form')
    expect(form.exists()).toBeTruthy()

    const textArea = form.find('#auto-signup-domains')
    const textAreaElement = textArea.element as HTMLTextAreaElement
    expect(textArea.exists()).toBeTruthy()
    expect(textAreaElement.value).toBe('test.com, google.com, networknext.com')

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks auto sign up form submission', async () => {
    let updateDomainsSpy = updateAutoSignupDomainsMock(localVue, true, ['test.com', 'google.com', 'networknext.com'])
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const form = wrapper.find('#auto-signup-form')
    expect(form.exists()).toBeTruthy()

    let textArea = form.find('#auto-signup-domains')
    let textAreaElement = textArea.element as HTMLTextAreaElement
    expect(textArea.exists()).toBeTruthy()
    expect(textAreaElement.value).toBe('')

    await textArea.setValue('test.com, google.com, networknext.com')

    expect(textAreaElement.value).toBe('test.com, google.com, networknext.com')
    expect(wrapper.vm.$data.autoSignupDomains).toBe('test.com, google.com, networknext.com')

    const button = form.find('#auto-signup-button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Save Automatic Sign up')

    await button.trigger('submit')
    await localVue.nextTick()

    expect(updateDomainsSpy).toBeCalledTimes(1)

    let alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()
    expect(alert.text()).toBe('Successfully updated signup domains')

    expect(store.getters.userProfile.domains.length).toBe(3)
    expect(store.getters.userProfile.domains[0]).toBe('test.com')
    expect(store.getters.userProfile.domains[1]).toBe('google.com')
    expect(store.getters.userProfile.domains[2]).toBe('networknext.com')

    textArea = form.find('#auto-signup-domains')
    textAreaElement = textArea.element as HTMLTextAreaElement
    expect(textArea.exists()).toBeTruthy()
    expect(textAreaElement.value).toBe('test.com, google.com, networknext.com')

    await textArea.setValue('test.com')

    expect(textAreaElement.value).toBe('test.com')
    expect(wrapper.vm.$data.autoSignupDomains).toBe('test.com')

    await button.trigger('submit')
    await localVue.nextTick()

    expect(updateDomainsSpy).toBeCalledTimes(2)

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()
    expect(alert.text()).toBe('Successfully updated signup domains')

    expect(store.getters.userProfile.domains.length).toBe(1)
    expect(store.getters.userProfile.domains[0]).toBe('test.com')

    updateDomainsSpy = updateAutoSignupDomainsMock(localVue, false, [])

    expect(textAreaElement.value).toBe('test.com')
    expect(wrapper.vm.$data.autoSignupDomains).toBe('test.com')

    await button.trigger('submit')
    await localVue.nextTick()

    expect(updateDomainsSpy).toBeCalledTimes(3)

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()

    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe('Failed to update signup domains')

    expect(store.getters.userProfile.domains.length).toBe(1)
    expect(store.getters.userProfile.domains[0]).toBe('test.com')

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    updateDomainsSpy.mockReset()
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks add new users form - empty', () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const form = wrapper.find('#new-user-form')
    expect(form.exists()).toBeTruthy()

    const labels = form.findAll('label')
    expect(labels.length).toBe(2)
    expect(labels.at(0).text()).toBe('Add users by email address')
    expect(labels.at(1).text()).toBe('Permission Level')

    const helperTexts = form.findAll('small')
    expect(helperTexts.length).toBe(2)
    expect(helperTexts.at(0).text()).toContain('Enter a newline or comma-delimited list of email')
    expect(helperTexts.at(0).text()).toContain('addresses to add users to your account.')
    expect(helperTexts.at(1).text()).toBe('The permission level to grant the added user accounts.')

    const button = form.find('#add-user-button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Add Users')

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks add new user form - !empty', async () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks add new user form - !empty - multiple', async () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks existing users manager', () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()

    wrapper.destroy()
  })
})

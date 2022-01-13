import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import UserManagement from '@/components/UserManagement.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { VueConstructor } from 'vue/types/umd'
import { AlertType } from '@/components/types/AlertTypes'
import { faCheck, faDownload, faPen, faTimes, faTrash } from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

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
    ) : Promise.reject(new Error('allRolesMock Mock Error'))
  })
}

function updateAutoSignupDomainsMock (vueInstance: VueConstructor<any>, success: boolean, domains: Array<string>): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'updateAutoSignupDomains').mockImplementation(() => {
    return success ? Promise.resolve(
      {
        domains: domains
      }
    ) : Promise.reject(new Error('updateAutoSignupDomainsMock Mock Error'))
  })
}

function addNewUserAccounts (vueInstance: VueConstructor<any>, success: boolean, emails: Array<string>, roles: Array<any>, newAccounts: Array<any>): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'addNewUserAccounts').mockImplementation((args: any) => {
    return success ? Promise.resolve(
      {
        accounts: newAccounts
      }
    ) : Promise.reject(new Error('addNewUserAccounts Mock Error'))
  })
}

function updateUserRolesMock (vueInstance: VueConstructor<any>, success: boolean, userID: string, roles: Array<any>): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'updateUserRoles').mockImplementation((args: any) => {
    expect(args.user_id).toBe(userID)

    return success ? Promise.resolve(
      {
        roles: roles
      }
    ) : Promise.reject(new Error('updateUserRoles Mock Error'))
  })
}

function deleteUserAccountMock (vueInstance: VueConstructor<any>, success: boolean, userID: string): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'deleteUserAccount').mockImplementation((args: any) => {
    expect(args.user_id).toBe(userID)

    return success ? Promise.resolve() : Promise.reject(new Error('deleteUserAccount Mock Error'))
  })
}

describe('UserManagement.vue', () => {
  const localVue = createLocalVue()

  const ICONS = [
    faCheck,
    faPen,
    faTimes,
    faTrash
  ]

  library.add(...ICONS)

  // Mount FontAwesomeIcons
  localVue.component('font-awesome-icon', FontAwesomeIcon)


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
  // TODO: This should definitely happen, especially for the user tables tests
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

  it('checks add new users form layout', () => {
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

  it('checks add new user form functionality - single', async () => {
    const roles = [
      {
        name: 'Owner',
        description: 'Allow access to customer account'
      },
      {
        name: 'Explorer',
        description: 'Allow access to data analytics'
      }
    ]
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, roles)
    const addNewUserAccountsSpy = addNewUserAccounts(localVue, true, ['test@test.com'], roles, [
      {
        email: 'test@test.com',
        roles: [roles[0]]
      }
    ])

    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const form = wrapper.find('#new-user-form')
    expect(form.exists()).toBeTruthy()

    const roleDropDown = form.find('#role-drop-down')
    expect(roleDropDown.exists()).toBeTruthy()

    const roleInput = roleDropDown.find('input')
    expect(roleInput.exists()).toBeTruthy()

    const button = form.find('#add-user-button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Add Users')

    const emailInput = form.find('textarea')
    const emailInputElement = emailInput.element as HTMLTextAreaElement
    expect(emailInput.exists()).toBeTruthy()

    await emailInput.setValue('test@test.com')

    await localVue.nextTick()

    expect(emailInputElement.value).toBe('test@test.com')

    await roleInput.trigger('click')

    const roleOptions = wrapper.findAll('.multiselect__element')
    expect(roleOptions.length).toBe(2)

    // TODO: Can't get automated selection working so this is not a full integration test

    // Manually add new user role
    wrapper.vm.$data.newUserRoles.push(roles[0])

    await localVue.nextTick()

    await button.trigger('submit')

    await localVue.nextTick()

    expect(addNewUserAccountsSpy).toBeCalledTimes(1)

    expect(wrapper.vm.$data.companyUsers.length).toBe(1)
    expect(wrapper.vm.$data.companyUsers[0].email).toBe('test@test.com')
    expect(wrapper.vm.$data.companyUsers[0].roles.length).toBe(1)
    expect(wrapper.vm.$data.companyUsers[0].roles[0].name).toBe('Owner')

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(2)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    addNewUserAccountsSpy.mockReset()
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks add new user form functionality - multiple', async () => {
    const roles = [
      {
        name: 'Owner',
        description: 'Allow access to customer account'
      },
      {
        name: 'Explorer',
        description: 'Allow access to data analytics'
      }
    ]
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, roles)
    const addNewUserAccountsSpy = addNewUserAccounts(localVue, true, ['test2@test.com'], roles, [
      {
        email: 'test@test.com',
        roles: [roles[1]]
      },
      {
        email: 'test2@test.com',
        roles: [roles[1]]
      }
    ])

    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const form = wrapper.find('#new-user-form')
    expect(form.exists()).toBeTruthy()

    const roleDropDown = form.find('#role-drop-down')
    expect(roleDropDown.exists()).toBeTruthy()

    const roleInput = roleDropDown.find('input')
    expect(roleInput.exists()).toBeTruthy()

    const button = form.find('#add-user-button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Add Users')

    const emailInput = form.find('textarea')
    const emailInputElement = emailInput.element as HTMLTextAreaElement
    expect(emailInput.exists()).toBeTruthy()

    await emailInput.setValue('test@test.com, test2@test.com')

    await localVue.nextTick()

    expect(emailInputElement.value).toBe('test@test.com, test2@test.com')

    await roleInput.trigger('click')

    const roleOptions = wrapper.findAll('.multiselect__element')
    expect(roleOptions.length).toBe(2)

    // TODO: Can't get automated selection working so this is not a full integration test

    // Manually add new user role
    wrapper.vm.$data.newUserRoles.push(roles[0])

    await localVue.nextTick()

    await button.trigger('submit')

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()
    expect(alert.text()).toBe('User account(s) added successfully')

    expect(addNewUserAccountsSpy).toBeCalledTimes(1)

    expect(wrapper.vm.$data.companyUsers.length).toBe(2)
    expect(wrapper.vm.$data.companyUsers[0].email).toBe('test@test.com')
    expect(wrapper.vm.$data.companyUsers[0].roles.length).toBe(1)
    expect(wrapper.vm.$data.companyUsers[0].roles[0].name).toBe('Explorer')
    expect(wrapper.vm.$data.companyUsers[1].email).toBe('test2@test.com')
    expect(wrapper.vm.$data.companyUsers[1].roles.length).toBe(1)
    expect(wrapper.vm.$data.companyUsers[1].roles[0].name).toBe('Explorer')

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(2)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    addNewUserAccountsSpy.mockReset()
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks add new user form functionality - failure', async () => {
    const roles = [
      {
        name: 'Owner',
        description: 'Allow access to customer account'
      },
      {
        name: 'Explorer',
        description: 'Allow access to data analytics'
      }
    ]
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, roles)
    const addNewUserAccountsSpy = addNewUserAccounts(localVue, false, ['test2@test.com'], roles, [])

    const store = new Vuex.Store(defaultStore)
    const wrapper = mount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const form = wrapper.find('#new-user-form')
    expect(form.exists()).toBeTruthy()

    const roleDropDown = form.find('#role-drop-down')
    expect(roleDropDown.exists()).toBeTruthy()

    const roleInput = roleDropDown.find('input')
    expect(roleInput.exists()).toBeTruthy()

    const button = form.find('#add-user-button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Add Users')

    const emailInput = form.find('textarea')
    const emailInputElement = emailInput.element as HTMLTextAreaElement
    expect(emailInput.exists()).toBeTruthy()

    await emailInput.setValue('test@test.com, test2@test.com')

    await localVue.nextTick()

    expect(emailInputElement.value).toBe('test@test.com, test2@test.com')

    await roleInput.trigger('click')

    const roleOptions = wrapper.findAll('.multiselect__element')
    expect(roleOptions.length).toBe(2)

    // TODO: Can't get automated selection working so this is not a full integration test

    // Manually add new user role
    wrapper.vm.$data.newUserRoles.push(roles[0])

    await localVue.nextTick()

    await button.trigger('submit')

    await localVue.nextTick()

    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe('Failed to add user account(s)')

    expect(addNewUserAccountsSpy).toBeCalledTimes(1)

    expect(wrapper.vm.$data.companyUsers.length).toBe(0)

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(2)

    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())
    addNewUserAccountsSpy.mockReset()
    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })


  it('checks existing users manager - empty', () => {
    const allAccountsSpy = allAccountsMock(localVue, true, [])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    const userTableTitle = wrapper.find('#user-table-title')
    expect(userTableTitle.exists()).toBeTruthy()
    expect(userTableTitle.text()).toBe('Manage existing users')

    const userTableDetails = wrapper.find('#user-table-details')
    expect(userTableDetails.exists()).toBeTruthy()
    expect(userTableDetails.text()).toBe('Manage the list of users that currently have access to your Network Next company account.')

    const userTable = wrapper.find('table')
    expect(userTable.exists()).toBeTruthy()

    const headers = userTable.findAll('th')
    expect(headers.length).toBe(3)
    expect(headers.at(0).text()).toBe('Email Address')
    expect(headers.at(1).text()).toBe('Permissions')
    expect(headers.at(2).text()).toBe('Actions')

    const dataRows = userTable.findAll('td')
    expect(dataRows.length).toBe(1)
    expect(dataRows.at(0).text()).toBe('There are no users assigned to your company.')

    wrapper.destroy()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
  })

  it('checks existing users manager - not empty - success', async () => {
    const roles = [
      {
        name: 'Owner',
        description: 'Allow access to customer account'
      },
      {
        name: 'Explorer',
        description: 'Allow access to data analytics'
      }
    ]

    const allAccountsSpy = allAccountsMock(localVue, true, [
      {
        user_id: '00000000',
        email: 'test@test.com',
        roles: roles[0]
      }
    ])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const updateUserRolesSpy = updateUserRolesMock(localVue, true, 'auth0|00000000', [roles[1]])

    const deleteUserAccountSpy = deleteUserAccountMock(localVue, true, 'auth0|00000000')

    const store = new Vuex.Store(defaultStore)

    const wrapper = mount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const userTableTitle = wrapper.find('#user-table-title')
    expect(userTableTitle.exists()).toBeTruthy()
    expect(userTableTitle.text()).toBe('Manage existing users')

    const userTableDetails = wrapper.find('#user-table-details')
    expect(userTableDetails.exists()).toBeTruthy()
    expect(userTableDetails.text()).toBe('Manage the list of users that currently have access to your Network Next company account.')

    const userTable = wrapper.find('table')
    expect(userTable.exists()).toBeTruthy()

    const headers = userTable.findAll('th')
    expect(headers.length).toBe(3)
    expect(headers.at(0).text()).toBe('Email Address')
    expect(headers.at(1).text()).toBe('Permissions')
    expect(headers.at(2).text()).toBe('Actions')

    const dataRows = userTable.findAll('td')
    expect(dataRows.length).toBe(4)

    expect(dataRows.at(0).text()).toBe('test@test.com')

    const roleSelector = dataRows.at(1).find('.multiselect')
    expect(roleSelector.exists()).toBeTruthy()

    const editDeleteButtons = dataRows.at(2).findAll('button')
    expect(editDeleteButtons.length).toBe(2)

    const submitCancelButtons = dataRows.at(3).findAll('button')
    expect(submitCancelButtons.length).toBe(2)

    expect(dataRows.at(2).element.style.display).toBe('')
    expect(dataRows.at(3).element.style.display).toBe('none')

    expect(editDeleteButtons.at(0).attributes('id')).toBe('edit-user-button')
    expect(editDeleteButtons.at(1).attributes('id')).toBe('delete-user-button')

    const editButton = editDeleteButtons.at(0)
    const deleteButton = editDeleteButtons.at(1)

    await editButton.trigger('click')

    expect(wrapper.vm.$data.companyUsers[0].edit).toBeTruthy()
    expect(wrapper.vm.$data.companyUsers[0].delete).toBeFalsy()

    expect(dataRows.at(2).element.style.display).toBe('none')
    expect(dataRows.at(3).element.style.display).toBe('')

    expect(submitCancelButtons.at(0).attributes('id')).toBe('save-user-button')
    expect(submitCancelButtons.at(1).attributes('id')).toBe('cancel-user-button')

    const submitButton = submitCancelButtons.at(0)
    const cancelButton = submitCancelButtons.at(1)

    await cancelButton.trigger('click')

    expect(dataRows.at(2).element.style.display).toBe('')
    expect(dataRows.at(3).element.style.display).toBe('none')

    expect(wrapper.vm.$data.companyUsers[0].edit).toBeFalsy()
    expect(wrapper.vm.$data.companyUsers[0].delete).toBeFalsy()

    expect(dataRows.at(2).element.style.display).toBe('')
    expect(dataRows.at(3).element.style.display).toBe('none')

    await deleteButton.trigger('click')

    expect(wrapper.vm.$data.companyUsers[0].edit).toBeFalsy()
    expect(wrapper.vm.$data.companyUsers[0].delete).toBeTruthy()

    expect(dataRows.at(2).element.style.display).toBe('none')
    expect(dataRows.at(3).element.style.display).toBe('')

    await cancelButton.trigger('click')

    expect(dataRows.at(2).element.style.display).toBe('')
    expect(dataRows.at(3).element.style.display).toBe('none')

    expect(wrapper.vm.$data.companyUsers[0].edit).toBeFalsy()
    expect(wrapper.vm.$data.companyUsers[0].delete).toBeFalsy()

    expect(roleSelector.classes('multiselect--disabled')).toBeTruthy()

    await editButton.trigger('click')

    expect(roleSelector.attributes('disabled')).toBeFalsy()

    // TODO: Automatically selecting / removing roles isn't working
    wrapper.vm.$data.companyUsers[0].roles = [roles[1]]

    expect(submitButton.exists()).toBeTruthy()

    await submitButton.trigger('click')

    await localVue.nextTick()

    expect(updateUserRolesSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(allRolesSpy).toBeCalledTimes(2)

    let alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()
    expect(alert.text()).toBe('User account edited successfully')

    expect(wrapper.vm.$data.companyUsers[0].edit).toBeFalsy()
    expect(wrapper.vm.$data.companyUsers[0].delete).toBeFalsy()

    await deleteButton.trigger('click')

    await submitButton.trigger('click')

    await localVue.nextTick()

    expect(deleteUserAccountSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(allRolesSpy).toBeCalledTimes(3)

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()
    expect(alert.text()).toBe('User account deleted successfully')

    expect(wrapper.vm.$data.companyUsers.length).toBe(0)

    wrapper.destroy()

    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
    updateUserRolesSpy.mockReset()
    deleteUserAccountSpy.mockReset()
  })

  it('checks existing users manager - not empty - failure', async () => {
    const roles = [
      {
        name: 'Owner',
        description: 'Allow access to customer account'
      },
      {
        name: 'Explorer',
        description: 'Allow access to data analytics'
      }
    ]

    const allAccountsSpy = allAccountsMock(localVue, true, [
      {
        user_id: '00000000',
        email: 'test@test.com',
        roles: roles[0]
      }
    ])
    const allRolesSpy = allRolesMock(localVue, true, [])

    const updateUserRolesSpy = updateUserRolesMock(localVue, false, 'auth0|00000000', [roles[1]])

    const deleteUserAccountSpy = deleteUserAccountMock(localVue, false, 'auth0|00000000')

    const store = new Vuex.Store(defaultStore)

    const wrapper = mount(UserManagement, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    expect(allAccountsSpy).toBeCalledTimes(1)
    expect(allRolesSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const userTable = wrapper.find('table')
    expect(userTable.exists()).toBeTruthy()

    const dataRows = userTable.findAll('td')
    expect(dataRows.length).toBe(4)

    expect(dataRows.at(0).text()).toBe('test@test.com')

    const roleSelector = dataRows.at(1).find('.multiselect')
    expect(roleSelector.exists()).toBeTruthy()

    const editDeleteButtons = dataRows.at(2).findAll('button')
    expect(editDeleteButtons.length).toBe(2)

    const submitCancelButtons = dataRows.at(3).findAll('button')
    expect(submitCancelButtons.length).toBe(2)

    expect(dataRows.at(2).element.style.display).toBe('')
    expect(dataRows.at(3).element.style.display).toBe('none')

    expect(editDeleteButtons.at(0).attributes('id')).toBe('edit-user-button')
    expect(editDeleteButtons.at(1).attributes('id')).toBe('delete-user-button')

    const editButton = editDeleteButtons.at(0)
    const deleteButton = editDeleteButtons.at(1)
    const submitButton = submitCancelButtons.at(0)

    await editButton.trigger('click')

    expect(roleSelector.attributes('disabled')).toBeFalsy()
    expect(submitButton.exists()).toBeTruthy()

    await submitButton.trigger('click')

    await localVue.nextTick()

    expect(updateUserRolesSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(allRolesSpy).toBeCalledTimes(2)

    let alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe('Failed to edit user account')

    expect(wrapper.vm.$data.companyUsers[0].edit).toBeFalsy()
    expect(wrapper.vm.$data.companyUsers[0].delete).toBeFalsy()

    await deleteButton.trigger('click')

    await submitButton.trigger('click')

    await localVue.nextTick()

    expect(deleteUserAccountSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(allRolesSpy).toBeCalledTimes(3)

    alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe('Failed to delete user account')

    expect(wrapper.vm.$data.companyUsers.length).toBe(1)

    wrapper.destroy()

    allAccountsSpy.mockReset()
    allRolesSpy.mockReset()
    updateUserRolesSpy.mockReset()
    deleteUserAccountSpy.mockReset()
  })
})

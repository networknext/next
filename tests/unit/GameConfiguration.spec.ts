import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import GameConfiguration from '@/components/GameConfiguration.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { AlertType } from '@/components/types/AlertTypes'
import { UPDATE_PUBLIC_KEY_SUCCESS } from '@/components/types/Constants'
import { ErrorTypes } from '@/components/types/ErrorTypes'

describe('GameConfiguration.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultProfile = newDefaultProfile()

  const store = new Vuex.Store({
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
  })

  // Run bare minimum mount test
  it('mounts the component successfully', () => {
    const wrapper = shallowMount(GameConfiguration, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('mounts the component hidden - this state shouldn\'t be possible', () => {
    const wrapper = shallowMount(GameConfiguration, { localVue, store })

    expect(wrapper.exists()).toBeTruthy()

    // Main element should be hidden if the user isn't an Owner/Admin and doesn't have a company name
    expect(wrapper.find('div').exists()).toBeFalsy()

    wrapper.destroy()
  })

  it('checks elements - Owner - No pubkey', async () => {
    const newProfile = newDefaultProfile()
    newProfile.companyName = 'Test Company'
    newProfile.companyCode = 'test'
    newProfile.roles = ['Owner']

    store.commit('UPDATE_USER_PROFILE', newProfile)

    const wrapper = shallowMount(GameConfiguration, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeTruthy()

    // Allow the mounted function to update the UI
    await localVue.nextTick()

    // Check card information
    expect(wrapper.find('.card-title').text()).toBe('Game Configuration')
    expect(wrapper.find('.card-text').text()).toBe('Manage how your game connects to Network Next.')

    // Make sure the alert is hidden
    expect(wrapper.find('.alert').exists()).toBeFalsy()

    // Check labels
    const labels = wrapper.findAll('label')

    expect(labels.length).toBe(2)
    expect(labels.at(0).text()).toBe('Company Name')
    expect(labels.at(1).text()).toBe('Public Key')

    // Check company name display
    const companyNameInput = wrapper.find('#company-input')
    const companyNameInputElement = companyNameInput.element as HTMLInputElement

    expect(companyNameInput.exists()).toBeTruthy()
    expect(companyNameInputElement.value).toBe(store.getters.userProfile.companyName)
    expect(companyNameInput.attributes().disabled).toBe('disabled')

    // Check public key input - should be empty here
    const publicKeyInput = wrapper.find('#pubkey-input')
    const publicKeyInputElement = publicKeyInput.element as HTMLTextAreaElement

    expect(publicKeyInputElement.placeholder).toBe('Enter your base64-encoded public key')
    expect(publicKeyInputElement.value).toBe(store.getters.userProfile.pubKey)

    // Check save button - should be disabled
    const gameConfigButton = wrapper.find('#game-config-button')

    expect(gameConfigButton.text()).toBe('Save game configuration')
    expect(gameConfigButton.attributes().disabled).toBe('disabled')

    store.commit('UPDATE_USER_PROFILE', defaultProfile)

    wrapper.destroy()
  })

  it('checks elements - Owner - Pubkey', async () => {
    const newProfile = newDefaultProfile()
    newProfile.companyName = 'Test Company'
    newProfile.companyCode = 'test'
    newProfile.roles = ['Owner']
    newProfile.pubKey = btoa('test pubkey')

    store.commit('UPDATE_USER_PROFILE', newProfile)

    const wrapper = shallowMount(GameConfiguration, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeTruthy()

    // Allow the mounted function to update the UI
    await localVue.nextTick()

    // Check card information
    expect(wrapper.find('.card-title').text()).toBe('Game Configuration')
    expect(wrapper.find('.card-text').text()).toBe('Manage how your game connects to Network Next.')

    // Make sure the alert is hidden
    expect(wrapper.find('.alert').exists()).toBeFalsy()

    // Check labels
    const labels = wrapper.findAll('label')

    expect(labels.length).toBe(2)
    expect(labels.at(0).text()).toBe('Company Name')
    expect(labels.at(1).text()).toBe('Public Key')

    // Check company name display
    const companyNameInput = wrapper.find('#company-input')
    const companyNameInputElement = companyNameInput.element as HTMLInputElement

    expect(companyNameInput.exists()).toBeTruthy()
    expect(companyNameInputElement.value).toBe(store.getters.userProfile.companyName)
    expect(companyNameInput.attributes().disabled).toBe('disabled')

    // Check public key input - should not be empty here
    const publicKeyInput = wrapper.find('#pubkey-input')
    const publicKeyInputElement = publicKeyInput.element as HTMLTextAreaElement

    expect(publicKeyInputElement.value).toBe(store.getters.userProfile.pubKey)

    // Check save button - should not be disabled
    const gameConfigButton = wrapper.find('#game-config-button')

    expect(gameConfigButton.text()).toBe('Save game configuration')
    expect(gameConfigButton.attributes().disabled).toBeUndefined()

    store.commit('UPDATE_USER_PROFILE', defaultProfile)

    wrapper.destroy()
  })

  it('checks successful pubkey update', async () => {
    const updateGameConfigurationSpy = jest.spyOn(localVue.prototype.$apiService, 'updateGameConfiguration').mockImplementationOnce(() => {
      return Promise.resolve({
        game_config: {
          company: 'Test Company',
          buyer_id: '123456789',
          public_key: btoa('some random public key')
        }
      })
    })

    const fetchAllBuyersSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchAllBuyers').mockImplementationOnce(() => {
      return Promise.resolve({
        buyers: [
          {
            is_live: false,
            id: '123456789',
            company_code: 'test'
          }
        ]
      })
    })

    const spyPubKeyEntered = jest.spyOn(localVue.prototype.$apiService, 'sendPublicKeyEnteredSlackNotification').mockImplementation(() => {
      return Promise.resolve()
    })

    const newProfile = newDefaultProfile()
    newProfile.companyName = 'Test Company'
    newProfile.companyCode = 'test'
    newProfile.roles = ['Owner']

    store.commit('UPDATE_USER_PROFILE', newProfile)

    const wrapper = mount(GameConfiguration, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeTruthy()

    // Allow the mounted function to update the UI
    await localVue.nextTick()

    // Input public key
    const pubKeyTextArea = wrapper.find('#pubkey-input')
    const newPubKey = btoa('some random public key')

    await pubKeyTextArea.setValue(newPubKey)

    // Check public key input - should not be empty here
    const pubKeyTextAreaElement = pubKeyTextArea.element as HTMLTextAreaElement

    expect(pubKeyTextAreaElement.value).toBe(newPubKey)

    // Check save button - should not be disabled
    const gameConfigButton = wrapper.find('#game-config-button')

    expect(gameConfigButton.text()).toBe('Save game configuration')
    expect(gameConfigButton.attributes().disabled).toBeUndefined()

    // Check buyers list to make sure it is empty
    expect(store.getters.allBuyers.length).toBe(0)

    // Submit new public key
    await gameConfigButton.trigger('submit')

    // Check to make sure the spy functions were hit
    expect(updateGameConfigurationSpy).toBeCalledTimes(1)
    expect(spyPubKeyEntered).toBeCalledTimes(1)
    expect(fetchAllBuyersSpy).toBeCalledTimes(1)

    // Wait for UI to react
    await localVue.nextTick()

    // Check for alert
    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()
    expect(alert.text()).toBe(UPDATE_PUBLIC_KEY_SUCCESS)

    // Wait for all buyers call to finish
    await localVue.nextTick()

    // Check buyers list to make sure the new buyer was added correctly
    expect(store.getters.allBuyers.length).toBe(1)
    expect(store.getters.allBuyers[0].company_code).toBe('test')

    updateGameConfigurationSpy.mockReset()
    fetchAllBuyersSpy.mockReset()
    spyPubKeyEntered.mockReset()

    store.commit('UPDATE_USER_PROFILE', defaultProfile)
    store.commit('UPDATE_ALL_BUYERS', [])

    wrapper.destroy()
  })

  it('checks successful pubkey update - failed all buyers update', async () => {
    const updateGameConfigurationSpy = jest.spyOn(localVue.prototype.$apiService, 'updateGameConfiguration').mockImplementationOnce(() => {
      return Promise.resolve({
        game_config: {
          company: 'Test Company',
          buyer_id: '123456789',
          public_key: btoa('some random public key')
        }
      })
    })

    const fetchAllBuyersSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchAllBuyers').mockImplementationOnce(() => {
      return Promise.reject(new Error('Failed to fetch all buyers'))
    })

    const spyPubKeyEntered = jest.spyOn(localVue.prototype.$apiService, 'sendPublicKeyEnteredSlackNotification').mockImplementation(() => {
      return Promise.resolve()
    })

    const newProfile = newDefaultProfile()
    newProfile.companyName = 'Test Company'
    newProfile.companyCode = 'test'
    newProfile.roles = ['Owner']

    store.commit('UPDATE_USER_PROFILE', newProfile)

    const wrapper = mount(GameConfiguration, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeTruthy()

    // Allow the mounted function to update the UI
    await localVue.nextTick()

    // Input public key
    const pubKeyTextArea = wrapper.find('#pubkey-input')
    const newPubKey = btoa('some random public key')

    // Set the public key and wait for the reactive fields to update
    await pubKeyTextArea.setValue(newPubKey)

    // Check public key input - should not be empty here
    const pubKeyTextAreaElement = pubKeyTextArea.element as HTMLTextAreaElement

    expect(pubKeyTextAreaElement.value).toBe(newPubKey)

    // Check save button - should not be disabled
    const gameConfigButton = wrapper.find('#game-config-button')

    expect(gameConfigButton.text()).toBe('Save game configuration')
    expect(gameConfigButton.attributes().disabled).toBeUndefined()

    // Check buyers list to make sure it is empty
    expect(store.getters.allBuyers.length).toBe(0)

    // Submit new public key
    await gameConfigButton.trigger('submit')

    // Check to make sure the spy functions were hit
    expect(updateGameConfigurationSpy).toBeCalledTimes(1)
    expect(spyPubKeyEntered).toBeCalledTimes(1)
    expect(fetchAllBuyersSpy).toBeCalledTimes(1)

    // Wait for UI to react
    await localVue.nextTick()

    // Check for alert
    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.SUCCESS)).toBeTruthy()
    expect(alert.text()).toBe(UPDATE_PUBLIC_KEY_SUCCESS)

    // Check buyers list to make sure it is still empty
    expect(store.getters.allBuyers.length).toBe(0)

    updateGameConfigurationSpy.mockReset()
    fetchAllBuyersSpy.mockReset()
    spyPubKeyEntered.mockReset()

    store.commit('UPDATE_USER_PROFILE', defaultProfile)
    store.commit('UPDATE_ALL_BUYERS', [])

    wrapper.destroy()
  })

  it('checks failed pubkey update', async () => {
    const updateGameConfigurationSpy = jest.spyOn(localVue.prototype.$apiService, 'updateGameConfiguration').mockImplementationOnce(() => {
      return Promise.reject(new Error('Failed to update public key'))
    })

    const fetchAllBuyersSpy = jest.spyOn(localVue.prototype.$apiService, 'fetchAllBuyers').mockImplementationOnce(() => {
      return Promise.resolve({
        buyers: []
      })
    })

    const spyPubKeyEntered = jest.spyOn(localVue.prototype.$apiService, 'sendPublicKeyEnteredSlackNotification').mockImplementation(() => {
      return Promise.resolve()
    })

    const newProfile = newDefaultProfile()
    newProfile.companyName = 'Test Company'
    newProfile.companyCode = 'test'
    newProfile.roles = ['Owner']

    store.commit('UPDATE_USER_PROFILE', newProfile)

    const wrapper = mount(GameConfiguration, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeTruthy()

    // Allow the mounted function to update the UI
    await localVue.nextTick()

    // Input public key
    const pubKeyTextArea = wrapper.find('#pubkey-input')
    const newPubKey = btoa('some random public key')

    await pubKeyTextArea.setValue(newPubKey)

    // Check public key input - should not be empty here
    const pubKeyTextAreaElement = pubKeyTextArea.element as HTMLTextAreaElement

    expect(pubKeyTextAreaElement.value).toBe(newPubKey)

    // Check save button - should not be disabled
    const gameConfigButton = wrapper.find('#game-config-button')

    expect(gameConfigButton.text()).toBe('Save game configuration')
    expect(gameConfigButton.attributes().disabled).toBeUndefined()

    // Check buyers list to make sure it is empty
    expect(store.getters.allBuyers.length).toBe(0)

    // Submit new public key
    await gameConfigButton.trigger('submit')

    // Check to make sure the spy functions were hit
    expect(updateGameConfigurationSpy).toBeCalledTimes(1)
    expect(spyPubKeyEntered).toBeCalledTimes(0)
    expect(fetchAllBuyersSpy).toBeCalledTimes(0)

    // Wait for UI to react
    await localVue.nextTick()

    // Check for alert
    const alert = wrapper.find('.alert')
    expect(alert.exists()).toBeTruthy()
    expect(alert.classes(AlertType.ERROR)).toBeTruthy()
    expect(alert.text()).toBe(ErrorTypes.UPDATE_PUBLIC_KEY_FAILURE)

    // Check buyers list to make sure the new buyer was added correctly
    expect(store.getters.allBuyers.length).toBe(0)

    updateGameConfigurationSpy.mockReset()
    fetchAllBuyersSpy.mockReset()
    spyPubKeyEntered.mockReset()

    store.commit('UPDATE_USER_PROFILE', defaultProfile)
    store.commit('UPDATE_ALL_BUYERS', [])

    wrapper.destroy()
  })
})

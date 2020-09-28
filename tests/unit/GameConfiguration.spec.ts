import { shallowMount, createLocalVue, mount } from '@vue/test-utils'
import Vuex from 'vuex'
import GameConfiguration from '@/components/GameConfiguration.vue'
import { waitFor } from './utils'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { UserProfile } from '@/components/types/AuthTypes'

describe('GameConfiguration.vue', () => {

  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = new Vuex.Store({
    state: {
      userProfile: {
        company: ''
      }
    },
    getters: {
      userProfile: (state: any) => state.userProfile
    }
  })

  describe('GameConfiguration.vue', () => {
    it('mounts the game config tab in the settings workspace successfully', () => {
      const store = defaultStore
      const wrapper = shallowMount(GameConfiguration, { localVue, store })
      expect(wrapper.exists()).toBe(true)
    })

    it('checks to make sure all elements are correct', () => {
      const store = defaultStore
      const wrapper = shallowMount(GameConfiguration, { localVue, store })
      // Check card information
      expect(wrapper.find('.card-title').text()).toBe('Game Configuration')
      expect(wrapper.find('.card-text').text()).toBe('Manage how your game connects to Network Next.')

      // Make sure the alert is hidden
      expect(wrapper.find('.alert').exists()).toBe(false)

      // Check labels
      const labels = wrapper.findAll('label')
      expect(labels.length).toBe(2)
      expect(labels.at(0).text()).toBe('Company Name')
      expect(labels.at(1).text()).toBe('Public Key')
      wrapper.destroy()
    })

    it('checks state handling unauthorized', () => {
      const store = defaultStore
      const wrapper = shallowMount(GameConfiguration, { localVue, store })

      // Check inputs
      const input = wrapper.findAll('input')
      const textArea = wrapper.findAll('textarea')
      expect(input.length).toBe(1)
      expect(input.at(0).attributes('disabled')).toBe('disabled')
      expect(textArea.length).toBe(1)
      expect(input.at(0).attributes('disabled')).toBe('disabled')

      // Check button
      const button = wrapper.findAll('button')
      expect(button.length).toBe(0)
      wrapper.destroy()
    })

    it('checks state handling authorized', () => {
      const store = new Vuex.Store({
        state: {
          userProfile: {
            company: ''
          }
        },
        getters: {
          userProfile: (state: any) => state.userProfile,
          isOwner: () => true,
          isAdmin: () => false
        }
      })
      const wrapper = shallowMount(GameConfiguration, { localVue, store })
      // Check inputs
      const input = wrapper.findAll('input')
      const textArea = wrapper.findAll('textarea')
      expect(input.length).toBe(1)
      expect(input.at(0).attributes('disabled')).toBe("disabled")
      expect(textArea.length).toBe(1)
      expect(textArea.at(0).attributes('disabled')).toBe(undefined)

      // Check button
      const button = wrapper.findAll('button')
      expect(button.length).toBe(1)
      wrapper.destroy()
    })

    const spy = jest.spyOn(localVue.prototype.$apiService, 'updateGameConfiguration').mockImplementationOnce(() => {
      return Promise.resolve({
        game_config: {
          company: 'test company',
          buyer_id: '123456789',
          public_key: 'abcdefghijklmnopqrstuvwxyz'
        }
      })
    })

    const spy2 = jest.spyOn(localVue.prototype.$apiService, 'fetchAllBuyers').mockImplementationOnce(() => {
      return Promise.resolve({
        buyers: [
          {
            is_live: false,
            id: '1234',
            company_code: 'test'
          }
        ]
      })
    })

    /* it('checks the components response to the update game configuration call', async () => {
      const store = new Vuex.Store({
        state: {
          userProfile: {
            company: '',
            pubKey: '',
            buyerID: ''
          }
        },
        getters: {
          userProfile: (state: any) => state.userProfile,
          isOwner: () => true,
          isAdmin: () => false
        },
        mutations: {
          UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
            state.userProfile = userProfile
          },
        }
      })
      const wrapper = mount(GameConfiguration, { localVue, store })
      // Check inputs
      const input = wrapper.findAll('input')
      const textArea = wrapper.findAll('textarea')
      input.at(0).setValue('test company')
      textArea.at(0).setValue('abcdefghijklmnopqrstuvwxyz')

      wrapper.find('form').trigger('submit')

      expect(spy).toBeCalled()

      await waitFor(wrapper, '.alert')
      const alert = wrapper.find('.alert')
      expect(alert.exists()).toBe(true)
      expect(alert.classes().includes('alert-success')).toBe(true)
      expect(alert.text()).toBe('Updated public key successfully')

      expect(wrapper.vm.$store.state.userProfile.company).toBe('test company')
      expect(wrapper.vm.$store.state.userProfile.pubKey).toBe('abcdefghijklmnopqrstuvwxyz')
      expect(wrapper.vm.$store.state.userProfile.buyerID).toBe('123456789')
      wrapper.destroy()
    }) */
  })
})

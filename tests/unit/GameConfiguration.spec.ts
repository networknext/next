import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import GameConfiguration from '@/components/GameConfiguration.vue'
import { text } from '@fortawesome/fontawesome-svg-core'

describe('GameConfiguration.vue', () => {

  const localVue = createLocalVue()

  localVue.use(Vuex)

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
      expect(input.at(0).attributes('disabled')).toBe(undefined)
      expect(textArea.length).toBe(1)
      expect(input.at(0).attributes('disabled')).toBe(undefined)

      // Check button
      const button = wrapper.findAll('button')
      expect(button.length).toBe(1)
    })

    // TODO: Add in tests that check api response handling and the alert functionality
  })
})

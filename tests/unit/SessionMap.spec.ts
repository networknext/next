import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import SessionMap from '@/components/SessionMap.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

describe('UserManagement.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = new Vuex.Store({
    state: {
      userProfile: {
        companyCode: '',
        companyName: '',
        domains: []
      },
      currentFilter: {
        companyCode: ''
      }
    },
    getters: {
      userProfile: (state: any) => state.userProfile,
      currentFilter: (state: any) => state.currentFilter
    }
  })

  const spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchMapSessions').mockImplementation(() => {
    return Promise.resolve({
      map_points: [

      ]
    })
  })

  jest.useFakeTimers()

  describe('SessionMap.vue', () => {
    it('checks to see if everything is correct by default', () => {
      const store = defaultStore
      const wrapper = shallowMount(SessionMap, { localVue, store })
      expect(wrapper.exists()).toBe(true)
    })

    it('checks map points polling logic', () => {
      jest.clearAllMocks()
      const store = defaultStore
      const wrapper = shallowMount(SessionMap, { localVue, store })

      expect(spyMapPoints).toBeCalled()
      expect(spyMapPoints).toBeCalledTimes(1)
      jest.advanceTimersByTime(10000)
      // TODO: there is an issue here with how useFakeTimers() handles setInterval which duplicates the async calls within the interval
      // This isn't a huge issue now but it should be monitored so that the test can be corrected in the future
      // https://stackoverflow.com/questions/52177631/jest-timer-and-promise-dont-work-well-settimeout-and-async-function
      expect(spyMapPoints).toBeCalledTimes(3)
      jest.advanceTimersByTime(10000)
      expect(spyMapPoints).toBeCalledTimes(5)
      jest.advanceTimersByTime(10000)
      expect(spyMapPoints).toBeCalledTimes(7)
    })
  })
})

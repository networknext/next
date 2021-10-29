import { shallowMount, createLocalVue } from '@vue/test-utils'
import Vuex from 'vuex'
import SessionCounts from '@/components/SessionCounts.vue'
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
      },
      allBuyers: [
        {
          company_name: '',
          company_code: '',
          is_live: false
        }
      ],
      currentPage: 'map'
    },
    getters: {
      userProfile: (state: any) => state.userProfile,
      currentFilter: (state: any) => state.currentFilter,
      allBuyers: (state: any) => state.allBuyers,
      isAdmin: (state: any) => false,
      currentPage: (state: any) => state.currentPage
    }
  })

  const spyMapPoints = jest.spyOn(localVue.prototype.$apiService, 'fetchTotalSessionCounts').mockImplementation(() => {
    return Promise.resolve({
      direct: 0,
      next: 0
    })
  })

  jest.useFakeTimers()

  describe('SessionCounts.vue', () => {
    it('checks to see if everything is correct by default', () => {
      const store = defaultStore
      const wrapper = shallowMount(SessionCounts, { localVue, store })
      expect(wrapper.exists()).toBe(true)
    })

    it('checks count polling logic', () => {
      jest.clearAllMocks()
      const store = defaultStore
      const wrapper = shallowMount(SessionCounts, { localVue, store })

      expect(spyMapPoints).toBeCalled()
      expect(spyMapPoints).toBeCalledTimes(1)
      jest.advanceTimersByTime(1000)
      // TODO: there is an issue here with how useFakeTimers() handles setInterval which duplicates the async calls within the interval
      // This isn't a huge issue now but it should be monitored so that the test can be corrected in the future
      // https://stackoverflow.com/questions/52177631/jest-timer-and-promise-dont-work-well-settimeout-and-async-function
      expect(spyMapPoints).toBeCalledTimes(3)
      jest.advanceTimersByTime(1000)
      expect(spyMapPoints).toBeCalledTimes(5)
      jest.advanceTimersByTime(1000)
      expect(spyMapPoints).toBeCalledTimes(7)
    })

    it('checks to see if filters work on init', () => {
      const store = new Vuex.Store({
        state: {
          userProfile: {
            companyCode: 'test',
            companyName: 'Test Company',
            domains: []
          },
          currentFilter: {
            companyCode: ''
          },
          allBuyers: [
            {
              company_name: 'Test Company',
              company_code: 'test',
              is_live: true
            }
          ],
          currentPage: 'map'
        },
        getters: {
          userProfile: (state: any) => state.userProfile,
          currentFilter: (state: any) => state.currentFilter,
          allBuyers: (state: any) => state.allBuyers,
          isAdmin: (state: any) => false,
          currentPage: (state: any) => state.currentPage
        }
      })
      const wrapper = shallowMount(SessionCounts, { localVue, store })
    })
  })
})

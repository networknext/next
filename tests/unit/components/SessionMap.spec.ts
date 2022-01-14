import { shallowMount, createLocalVue } from '@vue/test-utils'
import SessionMap from '@/components/SessionMap.vue'
import Vuex from 'vuex'

describe('SessionMap.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)

  // Init the store instance
  const store = new Vuex.Store({
    state: {
      userProfile: {
        email: 'test@test.com',
        companyName: 'Test Company',
        companyCode: 'test'
      },
      finishedTours: [],
      isTour: false,
      viewport: null
    },
    getters: {
      isTour: (state: any) => state.isTour,
      finishedTours: (state: any) => state.finishedTours,
      isSignUpTour: () => false,
      userProfile: (state: any) => state.userProfile,
      currentViewport: (state: any) => state.viewport
    },
    actions: {
      updateCurrentViewport ({ commit }: any, viewport: any) {
        commit('UPDATE_CURRENT_VIEWPORT', viewport)
      }
    },
    mutations: {
      UPDATE_IS_TOUR (state: any, isTour: boolean) {
        state.isTour = isTour
      },
      UPDATE_CURRENT_VIEWPORT (state: any, viewport: any) {
        state.viewport = viewport
      }
    }
  })

  // Run bare minimum mount test
  it('mounts the map successfully', () => {
    const wrapper = shallowMount(SessionMap, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

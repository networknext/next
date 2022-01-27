import { shallowMount, createLocalVue } from '@vue/test-utils'
import MapWorkspace from '@/workspaces/MapWorkspace.vue'
import Vuex from 'vuex'

describe('MapWorkspace.vue', () => {
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
      isTour: false
    },
    getters: {
      isTour: (state: any) => state.isTour,
      finishedTours: (state: any) => state.finishedTours,
      isSignUpTour: () => false,
      userProfile: (state: any) => state.userProfile
    },
    mutations: {
      UPDATE_IS_TOUR (state: any, isTour: boolean) {
        state.isTour = isTour
      }
    }
  })

  // Run bare minimum mount test
  it('mounts the map workspace successfully', () => {
    const wrapper = shallowMount(MapWorkspace, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

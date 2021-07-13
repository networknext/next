import { shallowMount, createLocalVue } from '@vue/test-utils'
import DownloadsWorkspace from '@/workspaces/DownloadsWorkspace.vue'
import Vuex from 'vuex'
import {
  faDownload
} from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import VueTour from 'vue-tour'

describe('DownloadsWorkspace.vue', () => {
  const localVue = createLocalVue()

  const ICONS = [
    faDownload
  ]

  library.add(...ICONS)

  localVue.component('font-awesome-icon', FontAwesomeIcon)

  localVue.use(Vuex)

  const store = new Vuex.Store({
    state: {
    },
    getters: {
      isSignUpTour: () => false
    },
    mutations: {
    }
  })

  localVue.use(VueTour)

  it('mounts the downloads workspace successfully', () => {
    const wrapper = shallowMount(DownloadsWorkspace, { localVue, store })
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })

  it('checks if the links are correct', () => {
    const wrapper = shallowMount(DownloadsWorkspace, { localVue, store })
    expect(wrapper.find('.card-title').text()).toBe('Network Next SDK')

    expect(wrapper.findAll('.btn').length).toBe(3)
    expect(wrapper.findAll('.btn').at(0).text()).toBe('SDK v4.0.16')
    expect(wrapper.findAll('.btn').at(1).text()).toBe('UE4 Plugin')
    expect(wrapper.findAll('.btn').at(2).text()).toBe('Documentation')
    wrapper.destroy()
  })
})

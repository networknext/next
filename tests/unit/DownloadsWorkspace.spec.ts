import { shallowMount, createLocalVue } from '@vue/test-utils'
import DownloadsWorkspace from '@/workspaces/DownloadsWorkspace.vue'
import {
  faDownload
} from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

describe('Alert.vue', () => {

  const localVue = createLocalVue()

  const ICONS = [
    faDownload
  ]

  library.add(...ICONS)

  localVue.component('font-awesome-icon', FontAwesomeIcon)

  describe('DownloadsWorkspace.vue', () => {

    it('mounts the downloads workspace successfully', () => {
      const wrapper = shallowMount(DownloadsWorkspace, { localVue })
      expect(wrapper.exists()).toBe(true)
      wrapper.destroy()
    })

    it('checks if the links are correct', () => {
      const wrapper = shallowMount(DownloadsWorkspace, { localVue })
      expect(wrapper.find('.card-title').text()).toBe('Network Next SDK')

      expect(wrapper.findAll('.btn').length).toBe(2)
      expect(wrapper.findAll('.btn').at(0).text()).toBe('SDK v3.4.6')
      expect(wrapper.findAll('.btn').at(0).attributes('onclick'))
        .toBe("window.open('https://storage.googleapis.com/portal_sdk_download_storage_prod/next-3.4.6.zip')")
      expect(wrapper.findAll('.btn').at(1).text()).toBe('v3.4.6 Documentation')
      expect(wrapper.findAll('.btn').at(1).attributes('onclick'))
        .toBe("window.open('https://network-next-sdk.readthedocs-hosted.com/en/latest/')")
      wrapper.destroy()
    })
  })
})

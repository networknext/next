import { shallowMount, createLocalVue, Wrapper } from '@vue/test-utils'
import DownloadsWorkspace from '@/workspaces/DownloadsWorkspace.vue'
import Vuex from 'vuex'
import {
  faDownload
} from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import VueTour from 'vue-tour'
import { FlagPlugin } from '@/plugins/flags'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import VueGtag from 'vue-gtag'
import { CombinedVueInstance } from 'vue/types/vue'
import { FeatureEnum } from '@/components/types/FeatureTypes'
import { IMPORTANT_CLICKS_CATEGORY, SDK_DOCUMENTATION_EVENT, SDK_DOCUMENTATION_URL, SDK_DOWNLOAD_EVENT, SDK_DOWNLOAD_URL, UE4_PLUGIN_DOWNLOAD_EVENT, UE4_PLUGIN_DOWNLOAD_URL } from '@/components/types/Constants'

describe('DownloadsWorkspace.vue', () => {
  const localVue = createLocalVue()

  const ICONS = [
    faDownload
  ]

  library.add(...ICONS)

  // Mount FontAwesomeIcons
  localVue.component('font-awesome-icon', FontAwesomeIcon)

  // Setup plugins
  localVue.use(Vuex)
  localVue.use(FlagPlugin, {
    flags: [],
    useAPI: false,
    apiService: {}
  })
  localVue.use(JSONRPCPlugin)
  localVue.use(VueTour)
  localVue.use(VueGtag)

  // Init the store instance
  const store = new Vuex.Store({
    state: {
      userProfile: {
        email: 'test@test.com',
        companyName: 'Test Company',
        companyCode: 'test'
      }
    },
    getters: {
      isSignUpTour: () => false,
      userProfile: (state: any) => state.userProfile
    },
    mutations: {
    }
  })

  // Set up mocks for api service
  const apiService = localVue.prototype.$apiService

  // Init global expected variables
  let expectedURL = ''

  // Setup spy functions
  let windowSpy: jest.SpyInstance
  let spyDownloadSDK: jest.SpyInstance
  let spyDownloadUE4: jest.SpyInstance
  let spyDownloadDocs: jest.SpyInstance

  beforeEach(() => {
    windowSpy = jest.spyOn(window, 'window', 'get')
    windowSpy.mockImplementation(() => ({
      open: (url: string) => {
        expect(url).toBe(expectedURL)
      }
    }))
    spyDownloadSDK = jest.spyOn(apiService, 'sendSDKDownloadSlackNotification').mockImplementation(() => {})
    spyDownloadUE4 = jest.spyOn(apiService, 'sendUE4DownloadNotifications').mockImplementation(() => {})
    spyDownloadDocs = jest.spyOn(apiService, 'sendDocsViewSlackNotification').mockImplementation(() => {})
  })

  afterEach(() => {
    windowSpy.mockRestore()
    spyDownloadSDK.mockRestore()
    spyDownloadUE4.mockRestore()
    spyDownloadDocs.mockRestore()
  })

  // Run bare minimum mount test
  it('mounts the downloads workspace successfully', () => {
    const wrapper = shallowMount(DownloadsWorkspace, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  // Check the structure of the page
  it('checks the workspace structure', () => {
    const wrapper = shallowMount(DownloadsWorkspace, { localVue, store })
    const workspaceHeaders = wrapper.findAll('h1')
    expect(workspaceHeaders.length).toBe(1)
    expect(workspaceHeaders.at(0).text()).toBe('Downloads')
    expect(wrapper.find('.card-title').text()).toBe('Network Next SDK')

    const buttons = wrapper.findAll('.btn')

    expect(buttons.length).toBe(3)
    expect(buttons.at(0).text()).toBe('SDK v4.0.16')
    expect(buttons.at(1).text()).toBe('UE4 Plugin')
    expect(buttons.at(2).text()).toBe('Documentation')
  })

  // Check logic for button clicks
  it('checks button click logic without analytics', () => {
    const wrapper = shallowMount(DownloadsWorkspace, { localVue, store })
    const buttons = wrapper.findAll('.btn')

    expect(buttons.length).toBe(3)

    expectedURL = SDK_DOWNLOAD_URL

    buttons.at(0).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadSDK).toBeCalled()

    expectedURL = UE4_PLUGIN_DOWNLOAD_URL

    buttons.at(1).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadUE4).toBeCalled()

    expectedURL = SDK_DOCUMENTATION_URL

    buttons.at(2).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadDocs).toBeCalled()

    wrapper.destroy()
  })

  it('checks button click logic with analytics', () => {
    // Setup gtag mocks for analytics mocking
    const gtagService = localVue.prototype.$gtag
    localVue.prototype.$flagService.flags = [
      {
        name: FeatureEnum.FEATURE_ANALYTICS,
        description: 'Google analytics and tag manager hooks',
        value: true
      }
    ]

    const expectedCategory = IMPORTANT_CLICKS_CATEGORY
    let expectedEvent = ''
    const analyticsSpy = jest.spyOn(gtagService, 'event').mockImplementation((event: any, payload: any) => {
      expect(event).toBe(expectedEvent)
      expect(payload.event_category).toBe(expectedCategory)
    })

    const wrapper = shallowMount(DownloadsWorkspace, { localVue, store })
    const buttons = wrapper.findAll('.btn')
    expect(buttons.length).toBe(3)

    expectedURL = SDK_DOWNLOAD_URL
    expectedEvent = SDK_DOWNLOAD_EVENT

    buttons.at(0).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadSDK).toBeCalledTimes(1)

    expectedURL = UE4_PLUGIN_DOWNLOAD_URL
    expectedEvent = UE4_PLUGIN_DOWNLOAD_EVENT

    buttons.at(1).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadUE4).toBeCalledTimes(1)

    expectedURL = SDK_DOCUMENTATION_URL
    expectedEvent = SDK_DOCUMENTATION_EVENT

    buttons.at(2).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadDocs).toBeCalledTimes(1)

    analyticsSpy.mockReset()
  })
})
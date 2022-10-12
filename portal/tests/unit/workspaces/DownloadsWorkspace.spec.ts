import { shallowMount, createLocalVue } from '@vue/test-utils'
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
import { ENET_DOWNLOAD_EVENT, ENET_DOWNLOAD_URL, ENET_SOURCE_URL, SDK_DOCUMENTATION_URL, SDK_DOWNLOAD_URL, SDK_SOURCE_URL, UE4_PLUGIN_DOWNLOAD_URL, UE4_PLUGIN_SOURCE_URL, UNITY_PLUGIN_DOWNLOAD_URL, UNITY_PLUGIN_SOURCE_URL, WHITE_PAPER_DOWNLOAD_EVENT, WHITE_PAPER_DOWNLOAD_URL } from '@/components/types/Constants'

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
  let spyViewSDK: jest.SpyInstance
  let spyDownloadUE4: jest.SpyInstance
  let spyViewUE4: jest.SpyInstance
  let spyDownloadUnity: jest.SpyInstance
  let spyViewUnity: jest.SpyInstance
  let spyDownloadDocs: jest.SpyInstance
  let spyWhitePaperDownload: jest.SpyInstance
  let spyENetDownload: jest.SpyInstance
  let spyViewENet: jest.SpyInstance

  beforeEach(() => {
    windowSpy = jest.spyOn(window, 'window', 'get')
    windowSpy.mockImplementation(() => ({
      open: (url: string) => {
        expect(url).toBe(expectedURL)
      }
    }))
    spyDownloadSDK = jest.spyOn(apiService, 'sendSDKDownloadSlackNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyViewSDK = jest.spyOn(apiService, 'sendSDKSourceViewSlackNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyDownloadUE4 = jest.spyOn(apiService, 'sendUE4DownloadNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyViewUE4 = jest.spyOn(apiService, 'sendUE4SourceViewNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyDownloadUnity = jest.spyOn(apiService, 'sendUnityDownloadNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyViewUnity = jest.spyOn(apiService, 'sendUnitySourceViewNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyDownloadDocs = jest.spyOn(apiService, 'sendDocsViewSlackNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyWhitePaperDownload = jest.spyOn(apiService, 'send2022WhitePaperDownloadNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyENetDownload = jest.spyOn(apiService, 'sendENetDownloadNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyViewENet = jest.spyOn(apiService, 'sendENetSourceViewSlackNotification').mockImplementation(() => {
      return Promise.resolve()
    })
  })

  afterEach(() => {
    windowSpy.mockRestore()
    spyDownloadSDK.mockRestore()
    spyViewSDK.mockRestore()
    spyDownloadUE4.mockRestore()
    spyViewUE4.mockRestore()
    spyDownloadUnity.mockRestore()
    spyViewUnity.mockRestore()
    spyDownloadDocs.mockRestore()
    spyWhitePaperDownload.mockRestore()
    spyENetDownload.mockRestore()
    spyViewENet.mockRestore()
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

    expect(buttons.length).toBe(9)
    expect(buttons.at(0).text()).toBe('SDK v4.20')
    expect(buttons.at(1).text()).toBe('Github Source')
    expect(buttons.at(2).text()).toBe('Documentation')
    expect(buttons.at(3).text()).toBe('Plugin Download')
    expect(buttons.at(4).text()).toBe('Github Source')
    expect(buttons.at(5).text()).toBe('Plugin Download')
    expect(buttons.at(6).text()).toBe('Github Source')
    expect(buttons.at(7).text()).toBe('ENet Support')
    expect(buttons.at(8).text()).toBe('Github Source')
    // expect(buttons.at(4).text()).toBe('Download')
  })

  // Check logic for button clicks
  it('checks button click logic without analytics', () => {
    const wrapper = shallowMount(DownloadsWorkspace, { localVue, store })
    const buttons = wrapper.findAll('.btn')

    expect(buttons.length).toBe(9)

    expectedURL = SDK_DOWNLOAD_URL

    buttons.at(0).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadSDK).toBeCalled()

    expectedURL = SDK_SOURCE_URL

    buttons.at(1).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadSDK).toBeCalled()

    expectedURL = SDK_DOCUMENTATION_URL

    buttons.at(2).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadDocs).toBeCalled()

    expectedURL = UE4_PLUGIN_DOWNLOAD_URL

    buttons.at(3).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadUE4).toBeCalled()

    expectedURL = UE4_PLUGIN_SOURCE_URL

    buttons.at(4).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyViewUE4).toBeCalled()

    expectedURL = UNITY_PLUGIN_DOWNLOAD_URL

    buttons.at(5).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadUnity).toBeCalled()

    expectedURL = UNITY_PLUGIN_SOURCE_URL

    buttons.at(6).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyViewUnity).toBeCalled()

    expectedURL = ENET_DOWNLOAD_URL

    buttons.at(7).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyENetDownload).toBeCalled()

    expectedURL = ENET_SOURCE_URL

    buttons.at(8).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyENetDownload).toBeCalled()

    /*
      expectedURL = WHITE_PAPER_DOWNLOAD_URL

      buttons.at(4).trigger('click')
      expect(windowSpy).toBeCalled()
      expect(spyWhitePaperDownload).toBeCalled()
    */

    wrapper.destroy()
  })

  /** GTAG mock isn't working correctly when run in semaphore VM
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
  }) */
})

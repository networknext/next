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
import { ENET_DOWNLOAD_EVENT, ENET_DOWNLOAD_URL, SDK_DOCUMENTATION_URL, SDK_DOWNLOAD_URL, UE4_PLUGIN_DOWNLOAD_URL, WHITE_PAPER_DOWNLOAD_EVENT, WHITE_PAPER_DOWNLOAD_URL } from '@/components/types/Constants'

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
  let spyWhitePaperDownload: jest.SpyInstance
  let spyENetDownload: jest.SpyInstance

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
    spyDownloadUE4 = jest.spyOn(apiService, 'sendUE4DownloadNotifications').mockImplementation(() => {
      return Promise.resolve()
    })
    spyDownloadDocs = jest.spyOn(apiService, 'sendDocsViewSlackNotification').mockImplementation(() => {
      return Promise.resolve()
    })
    spyWhitePaperDownload = jest.spyOn(apiService, 'send2022WhitePaperDownloadNotifications').mockImplementation(() => {
      return Promise.resolve()
    })
    spyENetDownload = jest.spyOn(apiService, 'sendENetDownloadNotification').mockImplementation(() => {
      return Promise.resolve()
    })
  })

  afterEach(() => {
    windowSpy.mockRestore()
    spyDownloadSDK.mockRestore()
    spyDownloadUE4.mockRestore()
    spyDownloadDocs.mockRestore()
    spyWhitePaperDownload.mockRestore()
    spyENetDownload.mockRestore()
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

    expect(buttons.length).toBe(5)
    expect(buttons.at(0).text()).toBe('SDK v4.20')
    expect(buttons.at(1).text()).toBe('UE4 Plugin')
    expect(buttons.at(2).text()).toBe('ENet Support')
    expect(buttons.at(3).text()).toBe('Documentation')
    expect(buttons.at(4).text()).toBe('Download')
  })

  // Check logic for button clicks
  it('checks button click logic without analytics', () => {
    const wrapper = shallowMount(DownloadsWorkspace, { localVue, store })
    const buttons = wrapper.findAll('.btn')

    expect(buttons.length).toBe(5)

    expectedURL = SDK_DOWNLOAD_URL

    buttons.at(0).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadSDK).toBeCalled()

    expectedURL = UE4_PLUGIN_DOWNLOAD_URL

    buttons.at(1).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadUE4).toBeCalled()

    expectedURL = ENET_DOWNLOAD_URL

    buttons.at(2).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyENetDownload).toBeCalled()

    expectedURL = SDK_DOCUMENTATION_URL

    buttons.at(3).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyDownloadDocs).toBeCalled()

    expectedURL = WHITE_PAPER_DOWNLOAD_URL

    buttons.at(4).trigger('click')
    expect(windowSpy).toBeCalled()
    expect(spyWhitePaperDownload).toBeCalled()

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

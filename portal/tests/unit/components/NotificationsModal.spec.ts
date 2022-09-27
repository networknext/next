import Vuex from 'vuex'
import { createLocalVue, shallowMount } from '@vue/test-utils'
import NotificationsModal from '@/components/modals/NotificationsModal.vue'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { faChevronDown, faChevronLeft } from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { AuthPlugin } from '@/plugins/auth'
import { FeatureFlagService } from '@/plugins/flags'

function fetchNotificationsMock (
  vueInstance: VueConstructor<any>,
  success: boolean,
  releaseNotesNotifications: Array<any>,
  analyticsNotifications: Array<any>,
  invoiceNotifications: Array<any>,
  systemNotesNotifications: Array<any>
): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'fetchNotifications').mockImplementation(() => {
    return success ? Promise.resolve({
      release_notes_notifications: releaseNotesNotifications,
      analytics_notifications: analyticsNotifications,
      invoice_notifications: invoiceNotifications,
      system_notifications: systemNotesNotifications
    }) : Promise.reject(new Error('fetchNotificationsMock Mock Error'))
  })
}

function startAnalyticsTrialMock (
  vueInstance: VueConstructor<any>,
  success: boolean
): jest.SpyInstance<any, unknown[]> {
  return jest.spyOn(vueInstance.prototype.$apiService, 'startAnalyticsTrial').mockImplementation(() => {
    return success ? Promise.resolve() : Promise.reject(new Error('startAnalyticsTrialMock Mock Error'))
  })
}

describe('NotificationsModal.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const ICONS = [
    faChevronLeft,
    faChevronDown
  ]

  library.add(...ICONS)

  localVue.component('font-awesome-icon', FontAwesomeIcon)

  const defaultStore = {
    state: {
      isAdmin: false,
      isBuyer: true,
      isOwner: true
    },
    getters: {
      isAdmin: (state: any) => state.isAdmin,
      isBuyer: (state: any) => state.isBuyer,
      isOwner: (state: any) => state.isOwner
    }
  }

  it('mounts a notifications component successfully', () => {
    const spy = fetchNotificationsMock(localVue, true, [], [], [], [])

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(spy).toBeCalledTimes(1)

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks empty layout', () => {
    const spy = fetchNotificationsMock(localVue, true, [], [], [], [])

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(spy).toBeCalledTimes(1)

    const modalTitle = wrapper.find('.banner-message')
    expect(modalTitle.exists()).toBeTruthy()
    expect(modalTitle.text()).toBe('Account Notifications')

    const closeButton = wrapper.find('button')
    expect(closeButton.exists()).toBeTruthy()
    expect(closeButton.text()).toBe('Close')

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks release notes layout - single', async () => {
    const spy = fetchNotificationsMock(localVue, true, [
      {
        title: 'Test Notification'
      }
    ], [], [], [])

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    const notificationHeaders = wrapper.findAll('.card-header')
    expect(notificationHeaders.length).toBe(1)
    expect(notificationHeaders.at(0).text()).toBe('Test Notification')

    const notificationBody = wrapper.find('#release-notes-notification-0')
    expect(notificationBody.exists()).toBeTruthy()

    spy.mockReset()
    wrapper.destroy()
  })

  // TODO: test out multiple release notes

  it('checks release notes layout - multiple', async () => {
    const spy = fetchNotificationsMock(localVue, true, [
      {
        title: 'Test Notification'
      },
      {
        title: 'Test Notification 2'
      },
      {
        title: 'Test Notification 3'
      }
    ], [], [], [])

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    const notificationHeaders = wrapper.findAll('.card-header')
    expect(notificationHeaders.length).toBe(3)
    expect(notificationHeaders.at(0).text()).toBe('Test Notification')
    expect(notificationHeaders.at(1).text()).toBe('Test Notification 2')
    expect(notificationHeaders.at(2).text()).toBe('Test Notification 3')

    const notificationBody = wrapper.find('#release-notes-notification-0')
    expect(notificationBody.exists()).toBeTruthy()

    const notificationBody2 = wrapper.find('#release-notes-notification-1')
    expect(notificationBody2.exists()).toBeTruthy()

    const notificationBody3 = wrapper.find('#release-notes-notification-2')
    expect(notificationBody3.exists()).toBeTruthy()

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks release notes layout - failure', async () => {
    const spy = fetchNotificationsMock(localVue, false, [], [], [], [])

    const store = new Vuex.Store(defaultStore)
    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    const notificationHeaders = wrapper.findAll('.card-header')
    expect(notificationHeaders.length).toBe(0)

    spy.mockReset()
    wrapper.destroy()
  })

  it('checks analytics trial notification', async () => {
    const notificationsSpy = fetchNotificationsMock(localVue, true, [], [
      {
        title: 'Test Notification',
        message: 'Notification Message'
      }
    ], [], [])

    const analyticsTrialSpy = startAnalyticsTrialMock(localVue, true)

    const store = new Vuex.Store(defaultStore)

    localVue.use(AuthPlugin, {
      domain: 'domain',
      clientID: 'clientID',
      store: store,
      flagService: new FeatureFlagService({
        flags: [],
        useAPI: false
      })
    })

    const refreshTokenSpy = jest.spyOn(localVue.prototype.$authService, 'refreshToken').mockImplementation(() => {
      return new Promise((resolve: any) => {
        resolve()
      })
    })

    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(notificationsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const notificationHeaders = wrapper.findAll('.card-header')
    expect(notificationHeaders.length).toBe(1)
    expect(notificationHeaders.at(0).text()).toBe('Test Notification')

    const notificationBody = wrapper.find('#analytics-notification-0')
    expect(notificationBody.exists()).toBeTruthy()

    const notificationText = notificationBody.findAll('div')
    expect(notificationText.at(3).text()).toBe('Super cool analytics look :)') // TODO: Change this when everything is implemented fully
    expect(notificationText.at(4).text()).toBe('Notification Message')
    expect(notificationText.at(5).text()).toBe('Start Free Trial')

    const analyticsTrialButton = notificationBody.find('button')
    expect(analyticsTrialButton.exists()).toBeTruthy()
    expect(analyticsTrialButton.text()).toBe('Start Free Trial')

    await analyticsTrialButton.trigger('click')

    expect(analyticsTrialSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(refreshTokenSpy).toBeCalledTimes(1)

    refreshTokenSpy.mockReset()
    notificationsSpy.mockReset()
    analyticsTrialSpy.mockReset()
    wrapper.destroy()
  })

  it('checks analytics trial - fetch analytics failure', async () => {
    const notificationsSpy = fetchNotificationsMock(localVue, false, [], [], [], [])

    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(notificationsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const notificationHeaders = wrapper.findAll('.card-header')
    expect(notificationHeaders.length).toBe(0)

    notificationsSpy.mockReset()
    wrapper.destroy()
  })

  it('checks analytics trial - sign up failure', async () => {
    const notificationsSpy = fetchNotificationsMock(localVue, true, [], [
      {
        title: 'Test Notification',
        message: 'Notification Message'
      }
    ], [], [])

    const analyticsTrialSpy = startAnalyticsTrialMock(localVue, false)

    const store = new Vuex.Store(defaultStore)

    localVue.use(AuthPlugin, {
      domain: 'domain',
      clientID: 'clientID',
      store: store,
      flagService: new FeatureFlagService({
        flags: [],
        useAPI: false
      })
    })

    const refreshTokenSpy = jest.spyOn(localVue.prototype.$authService, 'refreshToken').mockImplementation(() => {
      return new Promise((resolve: any) => {
        resolve()
      })
    })

    const wrapper = shallowMount(NotificationsModal, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(notificationsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    const notificationHeaders = wrapper.findAll('.card-header')
    expect(notificationHeaders.length).toBe(1)
    expect(notificationHeaders.at(0).text()).toBe('Test Notification')

    const notificationBody = wrapper.find('#analytics-notification-0')
    expect(notificationBody.exists()).toBeTruthy()

    const notificationText = notificationBody.findAll('div')
    expect(notificationText.at(3).text()).toBe('Super cool analytics look :)') // TODO: Change this when everything is implemented fully
    expect(notificationText.at(4).text()).toBe('Notification Message')
    expect(notificationText.at(5).text()).toBe('Start Free Trial')

    const analyticsTrialButton = notificationBody.find('button')
    expect(analyticsTrialButton.exists()).toBeTruthy()
    expect(analyticsTrialButton.text()).toBe('Start Free Trial')

    await analyticsTrialButton.trigger('click')

    expect(analyticsTrialSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(refreshTokenSpy).toBeCalledTimes(0)

    refreshTokenSpy.mockReset()
    notificationsSpy.mockReset()
    analyticsTrialSpy.mockReset()
    wrapper.destroy()
  })
})

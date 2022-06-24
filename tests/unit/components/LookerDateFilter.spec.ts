import Vuex from 'vuex'
import { createLocalVue, shallowMount } from '@vue/test-utils'
import LookerDateFilter from '@/components/LookerDateFilter.vue'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

describe('LookerDateFilter.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = {
    state: {
      filter: {
        companyCode: '',
        dateRange: DateFilterType.LAST_7
      },
      isAdmin: false,
      isOwner: false,
      isExplorer: false,
      hasAnalytics: false
    },
    getters: {
      currentFilter: (state: any) => state.filter,
      isOwner: (state: any) => state.isOwner,
      isExplorer: (state: any) => state.isExplorer,
      isAdmin: (state: any) => state.isAdmin,
      hasAnalytics: (state: any) => state.hasAnalytics
    },
    actions: {
      updateCurrentFilter ({ commit }: any, filter: Filter) {
        commit('UPDATE_CURRENT_FILTER', filter)
      }
    },
    mutations: {
      UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
        state.filter = newFilter
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      },
      UPDATE_IS_OWNER (state: any, isOwner: boolean) {
        state.isOwner = isOwner
      },
      UPDATE_IS_EXPLORER (state: any, isExplorer: boolean) {
        state.isExplorer = isExplorer
      },
      UPDATE_HAS_ANALYTICS (state: any, hasAnalytics: boolean) {
        state.hasAnalytics = hasAnalytics
      }
    }
  }

  // Run bare minimum mount test
  it('mounts the component successfully', async () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(LookerDateFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#date-filter')
    expect(filter.exists()).toBeTruthy()

    wrapper.destroy()
  })

  it('checks default option list', async () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(LookerDateFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#date-filter')
    expect(filter.exists()).toBeTruthy()

    const dropDownOptions = wrapper.findAll('option')
    expect(dropDownOptions.length).toBe(1)

    expect(dropDownOptions.at(0).text()).toBe('Last 7 Days')

    wrapper.destroy()
  })

  it('checks mid tier option list - owner', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_OWNER', true)

    const wrapper = shallowMount(LookerDateFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#date-filter')
    expect(filter.exists()).toBeTruthy()

    const dropDownOptions = wrapper.findAll('option')
    expect(dropDownOptions.length).toBe(3)

    expect(dropDownOptions.at(0).text()).toBe('Last 7 Days')
    expect(dropDownOptions.at(1).text()).toBe('Last 14 Days')
    expect(dropDownOptions.at(2).text()).toBe('Last 30 Days')

    store.commit('UPDATE_IS_OWNER', false)

    wrapper.destroy()
  })

  it('checks mid tier option list - explorer', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_EXPLORER', true)

    const wrapper = shallowMount(LookerDateFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#date-filter')
    expect(filter.exists()).toBeTruthy()

    const dropDownOptions = wrapper.findAll('option')
    expect(dropDownOptions.length).toBe(3)

    expect(dropDownOptions.at(0).text()).toBe('Last 7 Days')
    expect(dropDownOptions.at(1).text()).toBe('Last 14 Days')
    expect(dropDownOptions.at(2).text()).toBe('Last 30 Days')

    store.commit('UPDATE_IS_EXPLORER', false)

    wrapper.destroy()
  })

  it('checks top tier option list', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_EXPLORER', true)
    store.commit('UPDATE_HAS_ANALYTICS', true)

    const wrapper = shallowMount(LookerDateFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#date-filter')
    expect(filter.exists()).toBeTruthy()

    const dropDownOptions = wrapper.findAll('option')
    expect(dropDownOptions.length).toBe(5)

    expect(dropDownOptions.at(0).text()).toBe('Last 7 Days')
    expect(dropDownOptions.at(1).text()).toBe('Last 14 Days')
    expect(dropDownOptions.at(2).text()).toBe('Last 30 Days')
    expect(dropDownOptions.at(3).text()).toBe('Last 60 Days')
    expect(dropDownOptions.at(4).text()).toBe('Last 90 Days')

    store.commit('UPDATE_IS_EXPLORER', false)
    store.commit('UPDATE_HAS_ANALYTICS', false)

    wrapper.destroy()
  })

  it('checks filter updates', async () => {
    const store = new Vuex.Store(defaultStore)

    store.commit('UPDATE_IS_EXPLORER', true)
    store.commit('UPDATE_HAS_ANALYTICS', true)

    const wrapper = shallowMount(LookerDateFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#date-filter')
    expect(filter.exists()).toBeTruthy()

    const dropDownOptions = wrapper.findAll('option')
    expect(dropDownOptions.length).toBe(5)

    expect(dropDownOptions.at(0).text()).toBe('Last 7 Days')
    expect(dropDownOptions.at(1).text()).toBe('Last 14 Days')
    expect(dropDownOptions.at(2).text()).toBe('Last 30 Days')
    expect(dropDownOptions.at(3).text()).toBe('Last 60 Days')
    expect(dropDownOptions.at(4).text()).toBe('Last 90 Days')

    expect(store.getters.currentFilter.dateRange).toBe('7 days')

    await dropDownOptions.at(4).setSelected()

    await localVue.nextTick()

    expect(store.getters.currentFilter.dateRange).toBe('90 days')

    await dropDownOptions.at(3).setSelected()

    await localVue.nextTick()

    expect(store.getters.currentFilter.dateRange).toBe('60 days')

    await dropDownOptions.at(2).setSelected()

    await localVue.nextTick()

    expect(store.getters.currentFilter.dateRange).toBe('30 days')

    await dropDownOptions.at(1).setSelected()

    await localVue.nextTick()

    expect(store.getters.currentFilter.dateRange).toBe('14 days')

    await dropDownOptions.at(0).setSelected()

    await localVue.nextTick()

    expect(store.getters.currentFilter.dateRange).toBe('7 days')

    store.commit('UPDATE_IS_EXPLORER', false)
    store.commit('UPDATE_HAS_ANALYTICS', false)

    wrapper.destroy()
  })
})
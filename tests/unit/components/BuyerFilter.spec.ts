import Vuex from 'vuex'
import { createLocalVue, shallowMount } from '@vue/test-utils'
import BuyerFilter from '@/components/BuyerFilter.vue'
import { newDefaultProfile, UserProfile } from '@/components/types/AuthTypes'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'

describe('BuyerFilter.vue', () => {
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  const defaultStore = {
    state: {
      allBuyers: [],
      filter: {
        companyCode: '',
        dateRange: DateFilterType.CURRENT_MONTH
      },
      isAdmin: false,
      isBuyer: true,
      userProfile: newDefaultProfile()
    },
    getters: {
      allBuyers: (state: any) => state.allBuyers,
      currentFilter: (state: any) => state.filter,
      userProfile: (state: any) => state.userProfile,
      isBuyer: (state: any) => state.isBuyer,
      isAdmin: (state: any) => state.isAdmin
    },
    actions: {
      updateCurrentFilter ({ commit }: any, filter: Filter) {
        commit('UPDATE_CURRENT_FILTER', filter)
      }
    },
    mutations: {
      UPDATE_ALL_BUYERS (state: any, allBuyers: Array<any>) {
        state.allBuyers = allBuyers
      },
      UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
        state.filter = newFilter
      },
      UPDATE_USER_PROFILE (state: any, userProfile: UserProfile) {
        state.userProfile = userProfile
      },
      UPDATE_IS_BUYER (state: any, isBuyer: boolean) {
        state.isBuyer = isBuyer
      },
      UPDATE_IS_ADMIN (state: any, isAdmin: boolean) {
        state.isAdmin = isAdmin
      }
    }
  }

  // Run bare minimum mount test
  it('mounts the component successfully', async () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(BuyerFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#buyer-filter')
    expect(filter.exists()).toBeTruthy()

    wrapper.destroy()
  })

  it('checks filter list with 0 buyers', async () => {
    const store = new Vuex.Store(defaultStore)

    const wrapper = shallowMount(BuyerFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#buyer-filter')
    expect(filter.exists()).toBeTruthy()

    expect(wrapper.vm.$data.filterOptions.length).toBe(1)
    expect(wrapper.vm.$data.filterOptions[0].name).toBe('All')
    expect(wrapper.vm.$data.filterOptions[0].value).toBe('')

    const options = wrapper.findAll('option')
    expect(options.length).toBe(1)
    expect(options.at(0).text()).toBe('All')

    wrapper.destroy()
  })


  it('checks filter list - !admin', async () => {
    const store = new Vuex.Store(defaultStore)
    const profile = newDefaultProfile()
    profile.companyCode = 'test'
    store.commit('UPDATE_USER_PROFILE', profile)
    store.commit('UPDATE_ALL_BUYERS', [
      {
        company_name: 'Test Company',
        company_code: 'test',
      },
      {
        company_name: 'Test Company 2',
        company_code: 'test2',
      },
      {
        company_name: 'Test Company 3',
        company_code: 'test3',
      },
    ])

    const wrapper = shallowMount(BuyerFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#buyer-filter')
    expect(filter.exists()).toBeTruthy()

    expect(wrapper.vm.$data.filterOptions.length).toBe(2)
    expect(wrapper.vm.$data.filterOptions[0].name).toBe('All')
    expect(wrapper.vm.$data.filterOptions[0].value).toBe('')
    expect(wrapper.vm.$data.filterOptions[1].name).toBe('Test Company')
    expect(wrapper.vm.$data.filterOptions[1].value).toBe('test')

    const options = wrapper.findAll('option')
    expect(options.length).toBe(2)
    expect(options.at(0).text()).toBe('All')
    expect(options.at(1).text()).toBe('Test Company')

    store.commit('UPDATE_ALL_BUYERS', [])
    store.commit('UPDATE_USER_PROFILE', newDefaultProfile())

    wrapper.destroy()
  })

  it('checks filter list - admin', async () => {
    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)
    store.commit('UPDATE_ALL_BUYERS', [
      {
        company_name: 'Test Company',
        company_code: 'test',
        is_live: true
      },
      {
        company_name: 'Test Company 2',
        company_code: 'test2',
        is_live: false
      },
      {
        company_name: 'Test Company 3',
        company_code: 'test3',
        is_live: true
      },
    ])

    const wrapper = shallowMount(BuyerFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#buyer-filter')
    expect(filter.exists()).toBeTruthy()

    expect(wrapper.vm.$data.filterOptions.length).toBe(3)
    expect(wrapper.vm.$data.filterOptions[0].name).toBe('All')
    expect(wrapper.vm.$data.filterOptions[0].value).toBe('')
    expect(wrapper.vm.$data.filterOptions[1].name).toBe('Test Company')
    expect(wrapper.vm.$data.filterOptions[1].value).toBe('test')
    expect(wrapper.vm.$data.filterOptions[2].name).toBe('Test Company 3')
    expect(wrapper.vm.$data.filterOptions[2].value).toBe('test3')

    const options = wrapper.findAll('option')
    expect(options.length).toBe(3)
    expect(options.at(0).text()).toBe('All')
    expect(options.at(1).text()).toBe('Test Company')
    expect(options.at(2).text()).toBe('Test Company 3')

    store.commit('UPDATE_ALL_BUYERS', [])
    store.commit('UPDATE_IS_ADMIN', false)

    wrapper.destroy()
  })

  it('checks filter - admin - !all', async () => {
    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ADMIN', true)
    store.commit('UPDATE_ALL_BUYERS', [
      {
        company_name: 'Test Company',
        company_code: 'test',
        is_live: true
      },
      {
        company_name: 'Test Company 2',
        company_code: 'test2',
        is_live: true
      },
      {
        company_name: 'Test Company 3',
        company_code: 'test3',
        is_live: true
      },
    ])

    const wrapper = shallowMount(BuyerFilter, { localVue, store, propsData: { includeAll: false } })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#buyer-filter')
    expect(filter.exists()).toBeTruthy()

    expect(wrapper.vm.$data.filterOptions.length).toBe(3)
    expect(wrapper.vm.$data.filterOptions[0].name).toBe('Test Company')
    expect(wrapper.vm.$data.filterOptions[0].value).toBe('test')
    expect(wrapper.vm.$data.filterOptions[1].name).toBe('Test Company 2')
    expect(wrapper.vm.$data.filterOptions[1].value).toBe('test2')
    expect(wrapper.vm.$data.filterOptions[2].name).toBe('Test Company 3')
    expect(wrapper.vm.$data.filterOptions[2].value).toBe('test3')

    const options = wrapper.findAll('option')
    expect(options.length).toBe(3)
    expect(options.at(0).text()).toBe('Test Company')
    expect(options.at(1).text()).toBe('Test Company 2')
    expect(options.at(2).text()).toBe('Test Company 3')

    expect(store.getters.currentFilter.companyCode).toBe('test')

    store.commit('UPDATE_ALL_BUYERS', [])
    store.commit('UPDATE_IS_ADMIN', false)
    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '', dateRange: DateFilterType.CURRENT_MONTH })

    wrapper.destroy()
  })

  it('checks hidden filter', async () => {
    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_BUYER', false)

    const wrapper = shallowMount(BuyerFilter, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    await localVue.nextTick()

    const filter = wrapper.find('#buyer-filter')
    expect(filter.exists()).toBeFalsy()

    store.commit('UPDATE_IS_BUYER', true)

    wrapper.destroy()
  })
})

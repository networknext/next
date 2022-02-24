import { shallowMount, createLocalVue } from '@vue/test-utils'
import SessionMap from '@/components/SessionMap.vue'
import Vuex from 'vuex'
import { JSONRPCPlugin } from '@/plugins/jsonrpc'
import { VueConstructor } from 'vue/types/umd'
import { DateFilterType, Filter } from '@/components/types/FilterTypes'
import { newDefaultProfile } from '@/components/types/AuthTypes'
import { MAX_RETRIES } from '@/components/types/Constants'

function fetchMapSessionsMock (localVue: VueConstructor<any>, success: boolean, mapPoints: Array<any>, customerCode: string) {
  return jest.spyOn(localVue.prototype.$apiService, 'fetchMapSessions').mockImplementation((args: any) => {
    expect(args.company_code).toBe(customerCode)
    return success ? Promise.resolve({
      map_points: mapPoints
    }) : Promise.reject(new Error('fetchMapSessionsMock Error'))
  })
}

describe('SessionMap.vue', () => {
  jest.useFakeTimers()
  const localVue = createLocalVue()

  localVue.use(Vuex)
  localVue.use(JSONRPCPlugin)

  // Init the store instance
  const defaultStore = {
    state: {
      userProfile: newDefaultProfile(),
      filter: {
        companyCode: '',
        dateRange: DateFilterType.LAST_7
      },
      killLoops: false,
      isAnonymous: true,
      isAnonymousPlus: false,
      viewport: null
    },
    getters: {
      userProfile: (state: any) => state.userProfile,
      killLoops: (state: any) => state.killLoops,
      isAnonymous: (state: any) => state.isAnonymous,
      isAnonymousPlus: (state: any) => state.isAnonymousPlus,
      currentFilter: (state: any) => state.filter,
      currentViewport: (state: any) => state.viewport
    },
    actions: {
      toggleKillLoops ({ commit }: any, killLoops: boolean) {
        commit('TOGGLE_KILL_LOOPS', killLoops)
      },
      updateCurrentViewport ({ commit }: any, viewport: any) {
        commit('UPDATE_CURRENT_VIEWPORT', viewport)
      }
    },
    mutations: {
      UPDATE_CURRENT_FILTER (state: any, newFilter: Filter) {
        state.filter = newFilter
      },
      TOGGLE_KILL_LOOPS (state: any, killLoops: boolean) {
        state.killLoops = killLoops
      },
      UPDATE_IS_ANONYMOUS_PLUS (state: any, isAnonymousPlus: boolean) {
        state.isAnonymousPlus = isAnonymousPlus
      },
      UPDATE_IS_ANONYMOUS (state: any, isAnonymous: boolean) {
        state.isAnonymous = isAnonymous
      },
      UPDATE_CURRENT_VIEWPORT (state: any, viewport: any) {
        state.viewport = viewport
      }
    }
  }

  // Run bare minimum mount test
  it('mounts the map successfully', () => {
    const store = new Vuex.Store(defaultStore)
    const mapPointsSpy = fetchMapSessionsMock(localVue, true, [
      [0, 0, true, '00000000']
    ], '')

    const wrapper = shallowMount(SessionMap, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(mapPointsSpy).toBeCalledTimes(1)

    mapPointsSpy.mockReset()
    wrapper.destroy()
  })

  it('checks layers - anonymous', async () => {
    const store = new Vuex.Store(defaultStore)
    const mapPointsSpy = fetchMapSessionsMock(localVue, true, [
      [0, 0, true, '00000000'],
      [0, 0, true, '00000001'],
      [0, 0, false, '00000002'],
      [0, 0, false, '00000003']
    ], '')

    const wrapper = shallowMount(SessionMap, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(mapPointsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(wrapper.vm.$data.layers.length).toBe(1)

    mapPointsSpy.mockReset()
    wrapper.destroy()
  })

  it('checks layers - anonymous plus', async () => {
    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ANONYMOUS', false)
    store.commit('UPDATE_IS_ANONYMOUS_PLUS', true)

    const mapPointsSpy = fetchMapSessionsMock(localVue, true, [
      [0, 0, true, '00000000'],
      [0, 0, true, '00000001'],
      [0, 0, false, '00000002'],
      [0, 0, false, '00000003']
    ], '')

    const wrapper = shallowMount(SessionMap, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(mapPointsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(wrapper.vm.$data.layers.length).toBe(1)

    mapPointsSpy.mockReset()
    store.commit('UPDATE_IS_ANONYMOUS', true)
    store.commit('UPDATE_IS_ANONYMOUS_PLUS', false)
    wrapper.destroy()
  })

  it('checks layers - all filter', async () => {
    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ANONYMOUS', false)

    const mapPointsSpy = fetchMapSessionsMock(localVue, true, [
      [0, 0, true, '00000000'],
      [0, 0, true, '00000001'],
      [0, 0, false, '00000002'],
      [0, 0, false, '00000003']
    ], '')

    const wrapper = shallowMount(SessionMap, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(mapPointsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(wrapper.vm.$data.layers.length).toBe(1)

    mapPointsSpy.mockReset()
    store.commit('UPDATE_IS_ANONYMOUS', true)
    wrapper.destroy()
  })

  it('checks layers - multiple layers', async () => {
    const store = new Vuex.Store(defaultStore)
    store.commit('UPDATE_IS_ANONYMOUS', false)
    store.commit('UPDATE_CURRENT_FILTER', { companyCode: 'test', dateRange: DateFilterType.LAST_7 })

    const mapPointsSpy = fetchMapSessionsMock(localVue, true, [
      [0, 0, true, '00000000'],
      [0, 0, true, '00000001'],
      [0, 0, false, '00000002'],
      [0, 0, false, '00000003']
    ], 'test')

    const wrapper = shallowMount(SessionMap, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    expect(mapPointsSpy).toBeCalledTimes(1)

    await localVue.nextTick()

    expect(wrapper.vm.$data.layers.length).toBe(2)

    mapPointsSpy.mockReset()
    store.commit('UPDATE_IS_ANONYMOUS', true)
    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '', dateRange: DateFilterType.LAST_7 })
    wrapper.destroy()
  })

  it('checks filter change update', async () => {
    const store = new Vuex.Store(defaultStore)
    let spy = fetchMapSessionsMock(localVue, true, [], '')

    // Mount the component
    const wrapper = shallowMount(SessionMap, { localVue, store })

    // Make sure that the api calls are being mocked
    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    spy = fetchMapSessionsMock(localVue, true, [], 'test')

    const newFilter: Filter = { companyCode: 'test', dateRange: DateFilterType.LAST_7 }
    store.commit('UPDATE_CURRENT_FILTER', newFilter)

    await localVue.nextTick()

    expect(spy).toBeCalledTimes(2)

    store.commit('UPDATE_CURRENT_FILTER', { companyCode: '', dateRange: DateFilterType.LAST_7 })

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks kill loops', async () => {
    const store = new Vuex.Store(defaultStore)
    let spy = fetchMapSessionsMock(localVue, true, [], '')

    let wrapper = shallowMount(SessionMap, { localVue, store })
    expect(wrapper.exists()).toBeTruthy()

    store.commit('TOGGLE_KILL_LOOPS', true)

    await localVue.nextTick()

    expect(spy).toBeCalled()

    wrapper.destroy()

    spy.mockReset()

    spy = fetchMapSessionsMock(localVue, true, [], '')

    wrapper = shallowMount(SessionMap, { localVue, store })

    await localVue.nextTick()

    expect(spy).not.toBeCalled()

    store.commit('TOGGLE_KILL_LOOPS', false)

    spy.mockReset()

    wrapper.destroy()
  })

  it('checks failed api call', async () => {
    const store = new Vuex.Store(defaultStore)

    const spy = fetchMapSessionsMock(localVue, false, [], '')

    const wrapper = shallowMount(SessionMap, { localVue, store })

    await localVue.nextTick()

    expect(store.getters.killLoops).toBeFalsy()

    expect(spy).toBeCalledTimes(1)

    await localVue.nextTick()

    let retryCount = wrapper.vm.$data.retryCount
    expect(retryCount).toBe(1)

    for (let i = 2; i <= MAX_RETRIES; i++) {
      jest.advanceTimersByTime(3000 * retryCount)

      await localVue.nextTick()

      expect(spy).toBeCalledTimes(i)

      await localVue.nextTick()

      retryCount = wrapper.vm.$data.retryCount
      expect(retryCount).toBe(i)
    }

    expect(store.getters.killLoops).toBeTruthy()

    spy.mockReset()

    wrapper.destroy()
  })
})

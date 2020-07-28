import { shallowMount } from '@vue/test-utils'
import SessionMap from '@/components/SessionMap.vue'
import APIService from '@/services/api.service'
import Vue from 'vue'

describe('SessionMap.vue', () => {
  it('Loads session map component', () => {
    Vue.prototype.$apiService = new APIService()
    const wrapper = shallowMount(SessionMap, {})
    expect(wrapper.attributes().id.match('map-container'))
  })
})

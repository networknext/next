import { shallowMount } from '@vue/test-utils'
import MapPointsModal from '@/components/MapPointsModal.vue'

describe('MapPointsModal.vue', () => {
  it('mounts a get access modal successfully', () => {
    const wrapper = shallowMount(MapPointsModal)
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

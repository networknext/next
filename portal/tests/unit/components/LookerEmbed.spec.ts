import { shallowMount } from '@vue/test-utils'
import LookerEmbed from '@/components/LookerEmbed.vue'

describe('LookerEmbed.vue', () => {
  // Run bare minimum mount test
  it('mounts the component successfully', () => {
    const wrapper = shallowMount(LookerEmbed)
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

import { shallowMount } from '@vue/test-utils'
import GistEmbed from '@/components/GistEmbed.vue'

describe('GistEmbed.vue', () => {
  // Run bare minimum mount test
  it('mounts the component successfully', () => {
    const wrapper = shallowMount(GistEmbed)
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

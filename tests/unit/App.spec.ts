import { createLocalVue, shallowMount } from '@vue/test-utils'
import { JsonRPCPlugin } from '@/plugins/jsonrpc'
import App from '@/App.vue'

describe('App.vue', () => {

  const localVue = createLocalVue()
  localVue.use(JsonRPCPlugin)

  it('mounts the app and setups up the initial state', () => {
    const wrapper = shallowMount(App, { localVue })
  })
})

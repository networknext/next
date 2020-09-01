import { shallowMount, createLocalVue, mount, createWrapper } from '@vue/test-utils'
import UserToolWorkspace from '@/workspaces/UserToolWorkspace.vue'
import VueRouter from 'vue-router'
import { waitFor } from './utils'

describe('UserToolWorkspace.vue', () => {
  const localVue = createLocalVue()

  const defaultRouter = new VueRouter({
    routes: [
      {
        path: '/user-tool',
        name: 'user-tool',
        component: UserToolWorkspace
      }
    ]
  })

  localVue.use(VueRouter)

  it('mounts the user sessions table successfully', () => {
    const router = defaultRouter
    const wrapper = shallowMount(UserToolWorkspace, { localVue, router })
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })

  it('check no sessions for user', () => {
    const router = defaultRouter

    // Trigger the user sessions page through navigation
    router.push({ path: '/user-tool' })

    // Check if navigation worked
    expect(router.currentRoute.fullPath).toBe('/user-tool')

    // Mount the component
    const wrapper = mount(UserToolWorkspace, { localVue, router })

    // Check Title
    expect(wrapper.find('.h2').text()).toBe('User Tool')

    // Check label
    expect(wrapper.find('label').text()).toBe('User Hash')

    // Check for an input
    expect(wrapper.find('input').exists()).toBe(true)

    // Check input placeholder
    expect(wrapper.find('input').attributes('placeholder')).toBe('Enter a User Hash to view statistics')

    // Check button
    expect(wrapper.find('button').text()).toBe('View Sessions')

    // Check for an info alert
    expect(wrapper.find('.alert').text()).toBe('Please enter a User ID or Hash to view their sessions.')
    wrapper.destroy()
  })

  it('type into input and search', () => {
    // TODO: Add a test that inputs a hash and hits the button
  })
})

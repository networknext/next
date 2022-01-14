import { createLocalVue, shallowMount } from '@vue/test-utils'
import ResetPasswordModal from '@/components/ResetPasswordModal.vue'

describe('ResetPasswordModal.vue', () => {
  const localVue = createLocalVue()

  // Setup spy functions
  let windowSpy: jest.SpyInstance

  beforeEach(() => {
    windowSpy = jest.spyOn(window, 'window', 'get')
    windowSpy.mockImplementation(() => ({
      location: {
        origin: '127.0.0.1'
      }
    }))
  })

  afterEach(() => {
    windowSpy.mockRestore()
  })

  const $route = {
    path: '/reset-password',
    params: {
      pathMatch: ''
    },
    query: {
      redirectURI: ''
    }
  }

  const mocks = {
    $route,
    $router: {
      push: (newRoute: any) => {
        $route.path = newRoute.path
      }
    }
  }

  it('mounts a password reset modal successfully', () => {
    const wrapper = shallowMount(ResetPasswordModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

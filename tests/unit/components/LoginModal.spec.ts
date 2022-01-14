import { createLocalVue, shallowMount } from '@vue/test-utils'
import LoginModal from '@/components/LoginModal.vue'

describe('LoginModal.vue', () => {
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
    path: '/login',
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

  it('mounts a login modal successfully', () => {
    const wrapper = shallowMount(LoginModal, { localVue, mocks })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })
})

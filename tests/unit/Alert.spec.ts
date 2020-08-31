import { shallowMount } from '@vue/test-utils'
import Alert from '@/components/Alert.vue'
import { AlertTypes } from '@/components/types/AlertTypes'

describe('Alert.vue', () => {
  it('mounts an alert successfully', () => {
    const wrapper = shallowMount(Alert)
    expect(wrapper.exists()).toBe(true)
  })

  it('mounts an alert with a message', () => {
    const wrapper = shallowMount(Alert, {
      propsData: {
        message: 'This is a test'
      }
    })

    expect(wrapper.exists()).toBe(true)
    // TODO combine these some how
    expect(wrapper.find('div').classes('alert')).toBe(true)
    expect(wrapper.find('div').classes(AlertTypes.DEFAULT)).toBe(true)
    // ^^^
    expect(wrapper.find('div').text()).toBe('This is a test')
  })

  it('mounts an alert with a message and alert type', () => {
    const wrapper = shallowMount(Alert, {
      propsData: {
        message: 'This is still a test',
        alertType: AlertTypes.SUCCESS
      }
    })
    // TODO combine these some how
    expect(wrapper.find('div').classes('alert')).toBe(true)
    expect(wrapper.find('div').classes(AlertTypes.SUCCESS)).toBe(true)
    // ^^^
    expect(wrapper.find('div').text()).toBe('This is still a test')
  })

  it('mounts an alert with a link', () => {
    const wrapper = shallowMount(Alert, {
      propsData: {
        message: 'This is a test with a link: ',
        alertType: AlertTypes.INFO
      },
      slots: {
        default: '<a>I am a link!</a>'
      }
    })

    expect(wrapper.find('div').text()).toBe('This is a test with a link: \n  I am a link!')
    expect(wrapper.find('a').text()).toBe('I am a link!')
  })

  it('tests computed properties', () => {
    let message = 'This is a test'
    let alertType = AlertTypes.ERROR
    const wrapper = shallowMount(Alert, {
      propsData: {
        message: message,
        alertType: alertType
      }
    })

    expect((wrapper.vm as any).alertMessage).toBe(message)
    expect((wrapper.vm as any).className).toBe(alertType)

    message = 'This is also a test'
    alertType = AlertTypes.SUCCESS

    wrapper.setProps({
      message: message,
      alertType: alertType
    })

    expect((wrapper.vm as any).alertMessage).toBe(message)
    expect((wrapper.vm as any).className).toBe(alertType)
  })
})

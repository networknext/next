import { shallowMount } from '@vue/test-utils'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'

describe('Alert.vue', () => {
  it('mounts an alert successfully', () => {
    const wrapper = shallowMount(Alert)
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
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
    expect(wrapper.find('div').classes(AlertType.DEFAULT)).toBe(true)
    // ^^^
    expect(wrapper.find('div').text()).toBe('This is a test')
    wrapper.destroy()
  })

  it('mounts an alert with a message and alert type', () => {
    const wrapper = shallowMount(Alert, {
      propsData: {
        message: 'This is still a test',
        alertType: AlertType.SUCCESS
      }
    })
    // TODO combine these some how
    expect(wrapper.find('div').classes('alert')).toBe(true)
    expect(wrapper.find('div').classes(AlertType.SUCCESS)).toBe(true)
    // ^^^
    expect(wrapper.find('div').text()).toBe('This is still a test')
    wrapper.destroy()
  })

  it('mounts an alert with a link', () => {
    const wrapper = shallowMount(Alert, {
      propsData: {
        message: 'This is a test with a link: ',
        alertType: AlertType.INFO
      },
      slots: {
        default: '<a>I am a link!</a>'
      }
    })

    expect(wrapper.find('div').text()).toBe('This is a test with a link: \n  I am a link!')
    expect(wrapper.find('a').text()).toBe('I am a link!')
    wrapper.destroy()
  })

  it('tests computed properties', () => {
    let message = 'This is a test'
    let alertType = AlertType.ERROR
    const wrapper = shallowMount(Alert, {
      propsData: {
        message: message,
        alertType: alertType
      }
    })

    expect((wrapper.vm as any).alertMessage).toBe(message)
    expect((wrapper.vm as any).className).toBe(alertType)

    message = 'This is also a test'
    alertType = AlertType.SUCCESS

    wrapper.setProps({
      message: message,
      alertType: alertType
    })

    expect((wrapper.vm as any).alertMessage).toBe(message)
    expect((wrapper.vm as any).className).toBe(alertType)
    wrapper.destroy()
  })
})

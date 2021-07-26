import { shallowMount } from '@vue/test-utils'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'

describe('Alert.vue', () => {
  it('mounts an alert successfully', () => {
    const wrapper = shallowMount(Alert)
    expect(wrapper.exists()).toBe(true)
    wrapper.destroy()
  })

  it('mounts an alert with a message', async () => {
    const wrapper = shallowMount(Alert)

    wrapper.setData({
      message: 'This is a test',
      alertType: AlertType.DEFAULT
    })

    wrapper.vm.$mount()

    expect(wrapper.exists()).toBe(true)
    // TODO combine these some how
    expect(wrapper.find('div').classes('alert')).toBe(true)
    expect(wrapper.find('div').classes(AlertType.DEFAULT)).toBe(true)
    // ^^^
    expect(wrapper.find('div').text()).toBe('This is a test')
    wrapper.destroy()
  })

  it('mounts an alert with a message and alert type', () => {
    const wrapper = shallowMount(Alert)

    wrapper.setData({
      message: 'This is still a test',
      alertType: AlertType.SUCCESS
    })
    wrapper.vm.$mount()

    // TODO combine these some how
    expect(wrapper.find('div').classes('alert')).toBe(true)
    expect(wrapper.find('div').classes(AlertType.SUCCESS)).toBe(true)
    // ^^^
    expect(wrapper.find('div').text()).toBe('This is still a test')
    wrapper.destroy()
  })

  it('mounts an alert with a link', () => {
    const wrapper = shallowMount(Alert, {
      slots: {
        default: '<a>I am a link!</a>'
      }
    })

    wrapper.setData({
      message: 'This is a test with a link: ',
      alertType: AlertType.INFO
    })
    wrapper.vm.$mount()

    expect(wrapper.find('div').text()).toBe('This is a test with a link: \n  I am a link!')
    expect(wrapper.find('a').text()).toBe('I am a link!')
    wrapper.destroy()
  })

  it('tests computed properties', () => {
    const wrapper = shallowMount(Alert)

    wrapper.setData({
      message: 'This is a test',
      alertType: AlertType.ERROR
    })
    wrapper.vm.$mount()

    expect((wrapper.vm as any).alertMessage).toBe('This is a test')
    expect((wrapper.vm as any).className).toBe(AlertType.ERROR)

    wrapper.setData({
      message: 'This is also a test',
      alertType: AlertType.SUCCESS
    })

    expect((wrapper.vm as any).alertMessage).toBe('This is also a test')
    expect((wrapper.vm as any).className).toBe(AlertType.SUCCESS)
    wrapper.destroy()
  })
})

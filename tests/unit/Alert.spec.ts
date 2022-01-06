import { shallowMount } from '@vue/test-utils'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'

describe('Alert.vue', () => {
  // Run bare minimum mount test
  it('mounts the component successfully', () => {
    const wrapper = shallowMount(Alert)
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('mounts a default alert hidden', () => {
    const wrapper = shallowMount(Alert)

    expect(wrapper.exists()).toBeTruthy()

    // Main element should be hidden by default until a message is set
    expect(wrapper.find('div').exists()).toBeFalsy()

    wrapper.destroy()
  })

  it('show alert by setting message property', () => {
    const wrapper = shallowMount(Alert)

    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeFalsy()

    wrapper.setData({
      message: 'This is a test'
    })

    wrapper.vm.$mount()

    const mainDiv = wrapper.find('div')
    expect(mainDiv.exists()).toBeTruthy()
    expect(mainDiv.classes('alert')).toBeTruthy()
    expect(mainDiv.classes(AlertType.DEFAULT)).toBeTruthy()
    expect(mainDiv.text()).toBe('This is a test')
  })

  it('mounts an alert with a message and alert type', () => {
    const wrapper = shallowMount(Alert)

    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeFalsy()

    wrapper.setData({
      message: 'This is still a test',
      alertType: AlertType.SUCCESS
    })
    wrapper.vm.$mount()

    const mainDiv = wrapper.find('div')
    expect(mainDiv.exists()).toBeTruthy()
    expect(mainDiv.classes('alert')).toBeTruthy()
    expect(mainDiv.classes(AlertType.SUCCESS)).toBeTruthy()
    expect(mainDiv.text()).toBe('This is still a test')
    wrapper.destroy()
  })

  it('mounts an alert with a link', () => {
    const wrapper = shallowMount(Alert, {
      slots: {
        default: '<a>I am a link!</a>'
      }
    })

    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeFalsy()

    wrapper.setData({
      message: 'This is a test with a link: ',
      alertType: AlertType.INFO
    })
    wrapper.vm.$mount()

    const mainDiv = wrapper.find('div')
    expect(mainDiv.exists()).toBeTruthy()
    expect(mainDiv.classes('alert')).toBeTruthy()
    expect(mainDiv.classes(AlertType.INFO)).toBeTruthy()
    expect(mainDiv.text()).toBe('This is a test with a link: \n  I am a link!')
    expect(wrapper.find('a').text()).toBe('I am a link!')
    wrapper.destroy()
  })

  it('tests computed properties', () => {
    const wrapper = shallowMount(Alert)

    expect(wrapper.exists()).toBeTruthy()
    expect(wrapper.find('div').exists()).toBeFalsy()

    wrapper.setData({
      message: 'This is a test',
      alertType: AlertType.ERROR
    })
    wrapper.vm.$mount()

    const mainDiv = wrapper.find('div')
    expect(mainDiv.exists()).toBeTruthy()

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

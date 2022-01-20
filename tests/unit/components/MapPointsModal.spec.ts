import { createLocalVue, shallowMount } from '@vue/test-utils'
import MapPointsModal from '@/components/MapPointsModal.vue'
import { faCircle } from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

describe('MapPointsModal.vue', () => {
  const localVue = createLocalVue()

  const ICONS = [
    faCircle
  ]

  library.add(...ICONS)

  const stubs = [
    'router-link'
  ]

  // Mount FontAwesomeIcons
  localVue.component('font-awesome-icon', FontAwesomeIcon)
  it('mounts a get access modal successfully', () => {
    const wrapper = shallowMount(MapPointsModal, { localVue, stubs })
    expect(wrapper.exists()).toBeTruthy()
    wrapper.destroy()
  })

  it('checks modal with no points', () => {
    const wrapper = shallowMount(MapPointsModal, { localVue, stubs })
    expect(wrapper.exists()).toBeTruthy()

    const cardTitle = wrapper.find('.card-title')
    expect(cardTitle.exists()).toBeTruthy()
    expect(cardTitle.text()).toBe('Session IDs by location')

    const table = wrapper.find('table')
    expect(table.exists()).toBeTruthy()

    const headers = table.findAll('th')
    expect(headers.length).toBe(4)
    expect(headers.at(0).text()).toBe('')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('Latitude')
    expect(headers.at(3).text()).toBe('Longitude')

    const cells = table.findAll('td')
    expect(cells.length).toBe(1)
    expect(cells.at(0).text()).toBe('No points could be found at this zoom level. Please zoom in further for better results')

    const button = wrapper.find('button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Close')
    wrapper.destroy()
  })

  it('checks modal with multiple points', () => {
    const wrapper = shallowMount(MapPointsModal, {
      localVue,
      stubs,
      propsData: {
        points: [
          {
            source: [0, 0, true, '00000000']
          },
          {
            source: [1, 1, false, '00000001']
          },
          {
            source: [2, 2, true, '00000002']
          },
          {
            source: [3, 3, false, '00000003']
          },
          {
            source: [4, 4, true, '']
          }
        ]
      }
    })
    expect(wrapper.exists()).toBeTruthy()

    const cardTitle = wrapper.find('.card-title')
    expect(cardTitle.exists()).toBeTruthy()
    expect(cardTitle.text()).toBe('Session IDs by location')

    const table = wrapper.find('table')
    expect(table.exists()).toBeTruthy()

    const headers = table.findAll('th')
    expect(headers.length).toBe(4)

    expect(headers.at(0).text()).toBe('')
    expect(headers.at(1).text()).toBe('Session ID')
    expect(headers.at(2).text()).toBe('Latitude')
    expect(headers.at(3).text()).toBe('Longitude')

    const dataRows = table.findAll('tr td')
    expect(dataRows.length).toBe(20)

    expect(dataRows.at(0).find('#status').classes('text-success')).toBeTruthy()
    expect(dataRows.at(1).text()).toBe('00000000')
    expect(dataRows.at(2).text()).toBe('0')
    expect(dataRows.at(3).text()).toBe('0')
    expect(dataRows.at(4).find('#status').classes('text-primary')).toBeTruthy()
    expect(dataRows.at(5).text()).toBe('00000001')
    expect(dataRows.at(6).text()).toBe('1')
    expect(dataRows.at(7).text()).toBe('1')
    expect(dataRows.at(8).find('#status').classes('text-success')).toBeTruthy()
    expect(dataRows.at(9).text()).toBe('00000002')
    expect(dataRows.at(10).text()).toBe('2')
    expect(dataRows.at(11).text()).toBe('2')
    expect(dataRows.at(12).find('#status').classes('text-primary')).toBeTruthy()
    expect(dataRows.at(13).text()).toBe('00000003')
    expect(dataRows.at(14).text()).toBe('3')
    expect(dataRows.at(15).text()).toBe('3')
    expect(dataRows.at(16).find('#status').classes('text-success')).toBeTruthy()
    expect(dataRows.at(17).text()).toBe('Unavailable')
    expect(dataRows.at(18).text()).toBe('4')
    expect(dataRows.at(19).text()).toBe('4')

    const button = wrapper.find('button')
    expect(button.exists()).toBeTruthy()
    expect(button.text()).toBe('Close')
    wrapper.destroy()
  })
})

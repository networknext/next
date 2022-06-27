<template>
  <div>
    <div class="row" style="text-align: center;" v-show="showSpinner">
      <div class="col"></div>
      <div class="col">
        <div
          class="spinner-border"
          role="status"
          id="customers-spinner"
          style="margin:1rem;"
        >
          <span class="sr-only">Loading...</span>
        </div>
      </div>
      <div class="col"></div>
    </div>
    <v-tour name="sessionDetailsTour" :steps="sessionDetailsTourSteps" :options="sessionDetailsTourOptions" :callbacks="sessionDetailsTourCallbacks"></v-tour>
    <Alert ref="inputAlert"/>
    <div class="row" v-if="showDetails">
      <div class="col-12 col-lg-8">
        <div class="card mb-2">
          <div id="latency-card" class="card-header">
            <strong>
              Latency
            </strong>
            <div class="float-right">
              <span style="color: rgb(0, 109, 44);">
                — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                — Direct
              </span>
            </div>
          </div>
          <div class="card-body" data-tour="latencyGraph">
            <div id="latency-chart-1"></div>
          </div>
        </div>
        <div id="jitter-card" class="card mb-2">
          <div class="card-header">
            <strong>
              Jitter
            </strong>
            <div class="float-right">
              <span style="color: rgb(0, 109, 44);">
                — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                — Direct
              </span>
            </div>
          </div>
          <div class="card-body">
            <div id="jitter-chart-1"></div>
          </div>
        </div>
        <div id="pl-card" class="card mb-2">
          <div class="card-header">
            <strong>
              Packet Loss
            </strong>
            <div class="float-right">
              <span style="color: rgb(0, 109, 44);">
                  — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                  — Direct
              </span>
            </div>
          </div>
          <div class="card-body">
            <div id="packet-loss-chart-1"></div>
          </div>
        </div>
        <div id="bw-card" class="card mb-2">
          <div class="card-header">
            <strong>
              Bandwidth
            </strong>
            <div class="float-right">
              <span style="color: orange;">
                — Up
              </span>
              <span style="color: blue;">
                — Down
              </span>
            </div>
          </div>
          <div class="card-body">
            <div id="bandwidth-chart-1"></div>
          </div>
        </div>
      </div>
      <div class="col-12 col-lg-4">
        <div class="card mb-5">
          <div class="card-img-top">
            <div style="width: 100%; height: 40vh; margin: 0px; padding: 0px; position: relative;">
              <div id="session-tool-map"></div>
              <canvas id="session-tool-deck-canvas"></canvas>
            </div>
          </div>
          <div id="meta-panel" class="card-body">
            <div class="card-text">
              <dl>
                <dt>
                  Datacenter
                </dt>
                <dd>
                  <em>
                    {{ meta.datacenter_alias !== "" ? meta.datacenter_alias : meta.datacenter_name }}
                  </em>
                </dd>
                <dt>
                  ISP
                </dt>
                <dd>
                  <em>
                    {{ meta.location.isp != '' ? meta.location.isp : 'Unknown' }}
                  </em>
                </dd>
                <div v-if="(!$store.getters.isAnonymous && !$store.getters.isAnonymousPlus && getCustomerCode(meta.customer_id) === $store.getters.userProfile.companyCode) || $store.getters.isAdmin">
                  <dt>
                    User Hash
                  </dt>
                  <dd>
                    <router-link :to="`/user-tool/${meta.user_hash}`" class="text-dark">{{ meta.user_hash }}</router-link>
                  </dd>
                </div>
                <div v-if="(!$store.getters.isAnonymous && meta.customer_id === $store.getters.userProfile.buyerID) || $store.getters.isAdmin">
                  <dt>
                      IP Address
                  </dt>
                  <dd>
                      {{ meta.client_addr }}
                  </dd>
                </div>
                <dt>
                    Platform
                </dt>
                <dd>
                    {{ meta.platform }}
                </dd>
                <dt v-if="!$store.getters.isAnonymous">
                    Customer
                </dt>
                <dd v-if="!$store.getters.isAnonymous">
                    {{
                        getCustomerName(meta.customer_id)
                    }}
                </dd>
                <dt>
                  SDK Version
                </dt>
                <dd>
                  {{ meta.sdk }}
                </dd>
                <dt>
                  Connection Type
                </dt>
                <dd>
                  {{ meta.connection }}
                </dd>
                <dt v-if="$store.getters.isAdmin" style="padding-top: 20px;">
                    Route
                </dt>
                <table id="route-table" class="table table-sm mt-1" v-if="$store.getters.isAdmin">
                  <thead>
                    <tr>
                      <th style="width: 50%;">
                        Name
                      </th>
                      <th style="width: 50%;">
                        Role
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                        <td>
                            {{ meta.client_addr }}
                        </td>
                        <td>
                            <em>
                                User (Player)
                            </em>
                        </td>
                    </tr>
                    <tr v-for="(hop, index) in meta.hops" :key="index" scope="row">
                      <td>
                          {{ hop.name }}
                      </td>
                      <td>
                          Hop {{ index + 1 }}
                      </td>
                    </tr>
                    <tr>
                      <td>
                        {{ meta.server_addr }}
                      </td>
                      <td>
                          <em>Destination Server</em>
                      </td>
                    </tr>
                  </tbody>
                </table>
                <!-- TODO: Combine this so that we only check is Admin once -->
                <dt v-if="$store.getters.isAdmin && meta.nearby_relays.length > 0">
                    Nearby Relays
                </dt>
                <dd v-if="$store.getters.isAdmin && meta.nearby_relays.length === 0 && getBuyerIsLive(meta.customer_id)">
                  No Near Relays
                </dd>
                <dd v-if="$store.getters.isAdmin && meta.nearby_relays.length === 0 && !getBuyerIsLive(meta.customer_id)">
                  Customer is not live
                </dd>
                <table id="nearby-relays-table" class="table table-sm mt-1" v-if="$store.getters.isAdmin && meta.nearby_relays.length > 0">
                  <thead>
                    <tr>
                      <th style="width: 50%;">
                        Name
                      </th>
                      <th style="width: 16.66%;">
                        RTT
                      </th>
                      <th style="width: 16.66%;">
                        Jitter
                      </th>
                      <th style="width: 16.66%;">
                        Packet Loss
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                      <tr v-for="(relay, index) in meta.nearby_relays" :key="index">
                        <td>
                          <a class="text-dark">{{relay.name}}</a>&nbsp;
                        </td>
                        <td>
                          {{ parseFloat(relay.client_stats.rtt).toFixed(2) >= 255 ? '-' : parseFloat(relay.client_stats.rtt).toFixed(2) }}
                        </td>
                        <td>
                          {{ parseFloat(relay.client_stats.rtt).toFixed(2) >= 255 ? '-' : parseFloat(relay.client_stats.jitter).toFixed(2) }}
                        </td>
                        <td>
                          {{ parseFloat(relay.client_stats.rtt).toFixed(2) >= 255 ? '-' : parseFloat(relay.client_stats.packet_loss).toFixed(2) + '%' }}
                        </td>
                      </tr>
                  </tbody>
                </table>
              </dl>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">

import uPlot from 'uplot'

import { Route, NavigationGuardNext } from 'vue-router'
import { Component, Vue } from 'vue-property-decorator'

import Alert from '@/components/Alert.vue'
import { AlertType } from './types/AlertTypes'
import { FeatureEnum } from './types/FeatureTypes'

// import data1 from '../../test_data/session_details.json'

/**
 * This component displays all of the information related to the session
 *  tool page in the Portal and has all the associated logic and api calls
 */

/**
 * TODO: Cleanup template
 * TODO: Cleanup this whole component...
 */

@Component({
  components: {
    Alert
  }
})
export default class SessionDetails extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    inputAlert: Alert;
  }

  private showDetails = false
  private showSpinner: boolean

  private searchID: string

  private meta: any = null
  private slices: Array<any> = []

  // TODO: Replace these with a null uPlot instance some how
  private latencyComparisonChart: any
  private jitterComparisonChart: any
  private packetLossComparisonChart: any
  private bandwidthChart: any

  // TODO: Replace these with a null Deck/Mapbox instance some how
  private deckGlInstance: any = null
  private mapInstance: any = null

  private detailsLoop: any = null

  private unwatchFilter: any = null

  private viewState = {
    latitude: 0,
    longitude: 0,
    zoom: 2,
    pitch: 0,
    bearing: 0,
    minZoom: 0
  }

  private sessionDetailsTourSteps: Array<any>
  private sessionDetailsTourOptions: any
  private sessionDetailsTourCallbacks: any

  constructor () {
    super()
    this.searchID = ''
    // this.slices = (data1 as any).result.slices
    // this.meta = (data1 as any).result.meta

    this.sessionDetailsTourSteps = [
      {
        target: '[data-tour="latencyGraph"]',
        header: {
          title: 'Session Details'
        },
        content: 'Stats about a specific session can be viewed in this <strong>Session Tool</strong>. These are real-time improvements to latency, jitter, and packet loss.',
        params: {
          placement: 'right',
          enableScrolling: false
        }
      }
    ]

    this.sessionDetailsTourOptions = {
      labels: {
        buttonSkip: 'OK',
        buttonPrevious: 'BACK',
        buttonNext: 'NEXT',
        buttonStop: 'OK'
      }
    }

    this.sessionDetailsTourCallbacks = {
      onFinish: () => {
        this.$store.commit('UPDATE_FINISHED_TOURS', 'session-details')
        if (Vue.prototype.$flagService.isEnabled(FeatureEnum.FEATURE_ANALYTICS)) {
          Vue.prototype.$gtag.event('Session details tour finished', {
            event_category: 'Tours'
          })
        }
      }
    }

    this.showSpinner = true
  }

  private mounted () {
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        this.restartLoop()
      }
    )

    this.searchID = this.$route.params.pathMatch || ''
    if (this.searchID !== '') {
      this.restartLoop()
    }
  }

  private beforeDestroy () {
    clearInterval(this.detailsLoop)
    this.unwatchFilter()
  }

  private beforeRouteEnter (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    if (to.params.pathMatch === '') {
      next({ name: 'session-tool' })
      return
    }
    next()
  }

  // TODO: Move this somewhere with other helper functions
  private getCustomerName (buyerID: string) {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].id === buyerID) {
        return allBuyers[i].company_name
      }
    }
    return 'Private'
  }

  private getCustomerCode (buyerID: string) {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].id === buyerID) {
        return allBuyers[i].company_code
      }
    }
    return ''
  }

  private getBuyerIsLive (buyerID: string) {
    const allBuyers: Array<any> = this.$store.getters.allBuyers || []

    for (let i = 0; i < allBuyers.length; i++) {
      if (allBuyers[i].id === buyerID) {
        return allBuyers[i].is_live
      }
    }

    return false
  }

  private fetchSessionDetails () {
    this.$apiService.fetchSessionDetails({
      session_id: this.searchID,
      timeframe: this.$store.getters.currentFilter.dateRange || '',
      customer_code: this.$store.getters.isAdmin ? this.$store.getters.currentFilter.companyCode : this.$store.getters.userProfile.companyCode
    })
      .then((response: any) => {
        this.meta = response.meta
        this.slices = response.slices

        const enableRefresh = response.refresh || false
        if (enableRefresh && !this.detailsLoop) {
          this.detailsLoop = setInterval(() => {
            this.fetchSessionDetails()
          }, 10000)
        }

        this.meta.connection = this.meta.connection === 'wifi' ? 'Wi-Fi' : this.meta.connection.charAt(0).toUpperCase() + this.meta.connection.slice(1)

        if (!this.showDetails) {
          this.showDetails = true
        }

        setTimeout(() => {
          this.generateCharts()
          const NNCOLOR = [40, 167, 69]
          const DIRECTCOLOR = [49, 130, 189]

          const cellSize = 10
          const aggregation = 'MEAN'
          const gpuAggregation = false

          this.viewState.latitude = this.meta.location.latitude
          this.viewState.longitude = this.meta.location.longitude

          if (!this.mapInstance) {
            this.mapInstance = new (window as any).mapboxgl.Map({
              accessToken: process.env.VUE_APP_MAPBOX_TOKEN,
              style: 'mapbox://styles/mapbox/dark-v10',
              center: [
                this.meta.location.longitude,
                this.meta.location.latitude
              ],
              zoom: 2,
              pitch: 0,
              bearing: 0,
              container: 'session-tool-map'
            })
          }
          const sessionLocationLayer = new (window as any).deck.ScreenGridLayer({
            id: 'session-location-layer',
            data: [this.meta],
            opacity: 0.8,
            getPosition: (d: any) => [d.location.longitude, d.location.latitude],
            getWeight: () => 1,
            cellSizePixels: cellSize,
            colorRange: this.meta.on_network_next ? [NNCOLOR] : [DIRECTCOLOR],
            gpuAggregation,
            aggregation
          })

          if (!this.deckGlInstance) {
            // creating the deck.gl instance
            this.deckGlInstance = new (window as any).deck.Deck({
              canvas: document.getElementById('session-tool-deck-canvas'),
              width: '100%',
              height: '100%',
              initialViewState: this.viewState,
              controller: {
                dragRotate: false,
                dragTilt: false
              },
              // change the map's viewstate whenever the view state of deck.gl changes
              onViewStateChange: ({ viewState }: any) => {
                this.mapInstance.jumpTo({
                  center: [viewState.longitude, viewState.latitude],
                  zoom: viewState.zoom,
                  bearing: viewState.bearing,
                  pitch: viewState.pitch,
                  minZoom: 2,
                  maxZoom: 16
                })
              },
              layers: [sessionLocationLayer]
            })
          } else {
            this.deckGlInstance.setProps({ layers: [] })
            this.deckGlInstance.setProps({ layers: [sessionLocationLayer] })
          }
          if (this.$store.getters.isTour && this.$tours.sessionDetailsTour && !this.$tours.sessionDetailsTour.isRunning && !this.$store.getters.finishedTours.includes('session-details')) {
            this.$tours.sessionDetailsTour.start()
          }
        })
      })
      .catch((error: any) => {
        if (this.detailsLoop) {
          clearInterval(this.detailsLoop)
        }
        if (this.slices.length === 0) {
          console.log(`Something went wrong fetching sessions details for: ${this.searchID}`)
          console.log(error)
          this.$refs.inputAlert.setMessage('Failed to fetch session details')
          this.$refs.inputAlert.setAlertType(AlertType.ERROR)
        }
      })
      .finally(() => { this.showSpinner = false })
  }

  private restartLoop () {
    if (this.detailsLoop) {
      clearInterval(this.detailsLoop)
    }
    this.fetchSessionDetails()
  }

  private generateCharts () {
    const latencyData: Array<Array<number>> = [
      [],
      [],
      [],
      []
    ]
    const jitterData: Array<Array<number>> = [
      [],
      [],
      []
    ]
    const packetLossData: Array<Array<number>> = [
      [],
      [],
      []
    ]
    const bandwidthData: Array<Array<number>> = [
      [],
      [],
      []
    ]

    const latencyChartElement: HTMLElement | null = document.getElementById('latency-chart-1')

    const jitterChartElement: HTMLElement | null = document.getElementById('jitter-chart-1')

    const packetLossChartElement: HTMLElement | null = document.getElementById('packet-loss-chart-1')

    const bandwidthChartElement: HTMLElement | null = document.getElementById('bandwidth-chart-1')

    let directOnly = true

    this.slices.map((slice: any) => {
      const timestamp = new Date(slice.timestamp).getTime() / 1000
      const onNN = slice.on_network_next

      if (directOnly && onNN) {
        directOnly = false
      }

      const nextRTT = parseFloat(slice.next.rtt)
      const directRTT = parseFloat(slice.direct.rtt)

      const nextJitter = parseFloat(slice.next.jitter)
      const directJitter = parseFloat(slice.direct.jitter)

      const nextPL = parseFloat(slice.next.packet_loss)
      const directPL = parseFloat(slice.direct.packet_loss)

      // Latency
      let next = (slice.is_multipath && nextRTT >= directRTT && !this.$store.getters.isAdmin) ? directRTT : nextRTT
      next = (!this.$store.getters.isAdmin && slice.is_try_before_you_buy) ? 0 : next
      let direct = directRTT
      latencyData[0].push(timestamp)
      latencyData[1].push(next)
      latencyData[2].push(direct)
      latencyData[3].push(slice.predicted.rtt)

      // Jitter
      next = (slice.is_multipath && nextJitter >= directJitter && !this.$store.getters.isAdmin) ? directJitter : nextJitter
      next = (!this.$store.getters.isAdmin && slice.is_try_before_you_buy) ? 0 : next
      direct = directJitter
      jitterData[0].push(timestamp)
      jitterData[1].push(next)
      jitterData[2].push(direct)

      // Packetloss
      next = (slice.is_multipath && nextPL >= directPL && !this.$store.getters.isAdmin) ? directPL : nextPL
      next = (!this.$store.getters.isAdmin && slice.is_try_before_you_buy) ? 0 : next
      direct = directPL
      packetLossData[0].push(timestamp)
      packetLossData[1].push(next)
      packetLossData[2].push(direct)

      // Bandwidth
      bandwidthData[0].push(timestamp)
      bandwidthData[1].push(slice.envelope.up)
      bandwidthData[2].push(slice.envelope.down)
    })

    if (directOnly) {
      latencyData.splice(1, 1)
      jitterData.splice(1, 1)
      packetLossData.splice(1, 1)
    }

    let series = [
      {}
    ]

    if (!directOnly) {
      series.push({
        stroke: 'rgb(0, 109, 44)',
        fill: 'rgba(0, 109, 44, 0.1)',
        label: 'Network Next',
        value: (self: uPlot, rawValue: number) => rawValue.toFixed(2)
      })
    }

    series.push({
      stroke: 'rgb(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      label: 'Direct',
      value: (self: uPlot, rawValue: number) => rawValue.toFixed(2)
    })

    if (this.$store.getters.isAdmin) {
      series.push({
        stroke: 'rgb(255, 103, 0)',
        label: 'Predicted',
        value: (self: uPlot, rawValue: number) => rawValue.toFixed(2)
      })
    }

    let chartWidth = 0

    if (latencyChartElement) {
      chartWidth = latencyChartElement.clientWidth
    }

    const latencyComparisonOpts: uPlot.Options = {
      width: chartWidth,
      height: 260,
      cursor: {
        drag: {
          x: false,
          y: false
        }
      },
      scales: {
        // This causing the axis to not render correctly....
        /* ms: {
          from: 'y',
          auto: false,
          range: (self: uPlot, min: number, max: number): uPlot.MinMax => [
            0,
            max
          ]
        } */
      },
      series: series,
      axes: [
        {
          show: false
        },
        {
          scale: 'ms',
          show: true,
          gap: 5,
          size: 70,
          values: (self: uPlot, ticks: Array<number>) => ticks.map((rawValue: number) => rawValue + 'ms')
        }
      ]
    }

    series = [
      {}
    ]

    if (!directOnly) {
      series.push({
        stroke: 'rgb(0, 109, 44)',
        fill: 'rgba(0, 109, 44, 0.1)',
        label: 'Network Next',
        value: (self: uPlot, rawValue: number) => rawValue.toFixed(2)
      })
    }

    series.push({
      stroke: 'rgb(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      label: 'Direct',
      value: (self: uPlot, rawValue: number) => rawValue.toFixed(2)
    })

    const jitterComparisonOpts: uPlot.Options = {
      width: chartWidth,
      height: 260,
      cursor: {
        drag: {
          x: false,
          y: false
        }
      },
      scales: {
        // This causing the axis to not render correctly....
        /* ms: {
          from: 'y',
          auto: false,
          range: (self: uPlot, min: number, max: number): uPlot.MinMax => [
            0,
            max
          ]
        } */
      },
      series: series,
      axes: [
        {
          show: false
        },
        {
          scale: 'ms',
          show: true,
          gap: 5,
          size: 70,
          values: (self: uPlot, ticks: Array<number>) => ticks.map((rawValue: number) => rawValue + 'ms')
        }
      ]
    }

    series = [
      {}
    ]

    if (!directOnly) {
      series.push({
        stroke: 'rgb(0, 109, 44)',
        fill: 'rgba(0, 109, 44, 0.1)',
        label: 'Network Next',
        value: (self: uPlot, rawValue: number) => rawValue.toFixed(2)
      })
    }

    series.push({
      stroke: 'rgba(49, 130, 189)',
      fill: 'rgba(49, 130, 189, 0.1)',
      label: 'Direct',
      value: (self: uPlot, rawValue: number) => rawValue.toFixed(2)
    })

    const minMax: uPlot.MinMax = [0, 100]

    const packetLossComparisonOpts = {
      width: chartWidth,
      height: 260,
      cursor: {
        drag: {
          x: false,
          y: false
        }
      },
      scales: {
        y: {
          auto: false,
          range: minMax
        }
      },
      series: series,
      axes: [
        {
          show: false
        },
        {
          show: true,
          gap: 5,
          size: 50,
          values: (self: uPlot, ticks: Array<number>) => ticks.map((rawValue: number) => rawValue + '%')
        }
      ]
    }

    const bandwidthOpts = {
      width: chartWidth,
      height: 260,
      cursor: {
        drag: {
          x: false,
          y: false
        }
      },
      scales: {
        // This causing the axis to not render correctly....
        /* kbps: {
          from: 'y',
          auto: false,
          range: (self: uPlot, min: number, max: number): uPlot.MinMax => [
            0,
            max
          ]
        } */
      },
      series: [
        {},
        {
          stroke: 'orange',
          fill: 'rgba(255,165,0,0.1)',
          label: 'Up'
        },
        {
          stroke: 'blue',
          fill: 'rgba(0,0,255,0.1)',
          label: 'Down'
        }
      ],
      axes: [
        {
          show: false
        },
        {
          scale: 'kbps',
          show: true,
          gap: 5,
          size: 70,
          values: (self: uPlot, ticks: Array<number>) => ticks.map((rawValue: number) => rawValue + 'kbps')
        },
        {
          show: false
        }
      ]
    }

    if (this.latencyComparisonChart) {
      this.latencyComparisonChart.destroy()
    }

    if (latencyChartElement) {
      this.latencyComparisonChart = new uPlot(latencyComparisonOpts, latencyData, latencyChartElement)
    }

    if (this.jitterComparisonChart) {
      this.jitterComparisonChart.destroy()
    }

    if (jitterChartElement) {
      this.jitterComparisonChart = new uPlot(jitterComparisonOpts, jitterData, jitterChartElement)
    }

    if (this.packetLossComparisonChart) {
      this.packetLossComparisonChart.destroy()
    }

    if (packetLossChartElement) {
      this.packetLossComparisonChart = new uPlot(packetLossComparisonOpts, packetLossData, packetLossChartElement)
    }

    if (this.bandwidthChart) {
      this.bandwidthChart.destroy()
    }

    if (bandwidthChartElement) {
      this.bandwidthChart = new uPlot(bandwidthOpts, bandwidthData, bandwidthChartElement)
    }
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

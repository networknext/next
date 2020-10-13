<template>
  <div>
    <Alert :message="message" :alertType="alertType" v-if="message !== ''"/>
    <div class="row" v-if="showDetails">
      <div class="col-12 col-lg-8">
        <div class="card mb-2">
          <div class="card-header">
            <strong>
              Latency
            </strong>
            <div class="float-right">
              <span class="mr-2"
                    style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
              ></span>
              <span style="color: rgb(0, 109, 44);">
                — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                — Direct
              </span>
              <span></span>
            </div>
          </div>
          <div class="card-body">
            <div id="latency-chart-1"></div>
          </div>
        </div>
        <div class="card mb-2">
          <div class="card-header">
            <strong>
              Jitter
            </strong>
            <div class="float-right">
              <span class="mr-2"
                    style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
              ></span>
              <span style="color: rgb(0, 109, 44);">
                — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                — Direct
              </span>
              <span></span>
            </div>
          </div>
          <div class="card-body">
            <div id="jitter-chart-1"></div>
          </div>
        </div>
        <div class="card mb-2">
          <div class="card-header">
            <strong>
              Packet Loss
            </strong>
            <div class="float-right">
              <span class="mr-2"
                    style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
              ></span>
              <span style="color: rgb(0, 109, 44);">
                  — Network Next
              </span>
              <span style="color: rgb(49, 130, 189);">
                  — Direct
              </span>
              <span></span>
            </div>
          </div>
          <div class="card-body">
            <div id="packet-loss-chart-1"></div>
          </div>
        </div>
        <div class="card mb-2">
          <div class="card-header">
            <strong>
              Bandwidth
            </strong>
            <div class="float-right">
              <span class="mr-2"
                    style="border-right: 2px dotted rgb(51, 51, 51); display: none;"
              ></span>
              <span style="color: orange;">
                — Up
              </span>
              <span style="color: blue;">
                — Down
              </span>
              <span></span>
            </div>
            <div class="float-right">
              <span id="traffic-up-legend"></span>
              <span id="traffic-down-legend"></span>
            </div>
          </div>
          <div class="card-body">
            <div id="bandwidth-chart-1"></div>
          </div>
        </div>
      </div>
      <div class="col-12 col-lg-4">
        <div class="card">
          <div class="card-img-top">
            <div style="width: 100%; height: 40vh; margin: 0px; padding: 0px; position: relative;">
              <div id="session-tool-map"></div>
              <canvas id="session-tool-deck-canvas"></canvas>
            </div>
          </div>
          <div class="card-body">
            <div class="card-text">
              <dl>
                <dt>
                  ISP
                </dt>
                <dd>
                  <em>
                    {{ this.meta.location.isp != '' ? this.meta.location.isp : 'Unknown' }}
                  </em>
                </dd>
                <div v-if="!$store.getters.isAnonymous">
                  <dt>
                    User Hash
                  </dt>
                  <dd>
                    <router-link v-bind:to="`/user-tool/${this.meta.user_hash}`" class="text-dark">{{ this.meta.user_hash }}</router-link>
                  </dd>
                </div>
                <div v-if="(!$store.getters.isAnonymous && this.meta.buyer_id === $store.getters.userProfile.buyerID) || $store.getters.isAdmin">
                  <dt>
                      IP Address
                  </dt>
                  <dd>
                      {{ this.meta.client_addr }}
                  </dd>
                </div>
                <dt>
                    Platform
                </dt>
                <dd>
                    {{ this.meta.platform }}
                </dd>
                <dt v-if="!$store.getters.isAnonymous">
                    Customer
                </dt>
                <dd v-if="!$store.getters.isAnonymous">
                    {{
                        getCustomerName(this.meta.customer_id)
                    }}
                </dd>
                <dt>
                  SDK Version
                </dt>
                <dd>
                  {{ this.meta.sdk }}
                </dd>
                <dt>
                  Connection Type
                </dt>
                <dd>
                  {{ this.meta.connection }}
                </dd>
                <!-- TODO: Combine this so that we only check is Admin once -->
                <dt v-if="$store.getters.isAdmin && meta.nearby_relays.length > 0">
                    Nearby Relays
                </dt>
                <dd v-if="$store.getters.isAdmin && meta.nearby_relays.length == 0">
                    No Nearby Relays
                </dd>
                <table class="table table-sm mt-1" v-if="$store.getters.isAdmin && meta.nearby_relays.length > 0">
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
                      <tr v-for="(relay, index) in this.meta.nearby_relays" :key="index">
                        <td>
                          <a class="text-dark">{{relay.name}}</a>&nbsp;
                        </td>
                        <td>
                          {{ parseFloat(relay.client_stats.rtt).toFixed(2) }}
                        </td>
                        <td>
                          {{ parseFloat(relay.client_stats.jitter).toFixed(2) }}
                        </td>
                        <td>
                          {{ parseFloat(relay.client_stats.packet_loss).toFixed(2) }}%
                        </td>
                      </tr>
                  </tbody>
                </table>
                <dt  v-if="$store.getters.isAdmin">
                    Route
                </dt>
                <table class="table table-sm mt-1" v-if="$store.getters.isAdmin">
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
              </dl>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">

import mapboxgl from 'mapbox-gl'
import uPlot from 'uplot'

import { Deck } from '@deck.gl/core'
import { ScreenGridLayer } from '@deck.gl/aggregation-layers'
import { Route, NavigationGuardNext } from 'vue-router'
import { Component, Vue } from 'vue-property-decorator'

import 'uplot/dist/uPlot.min.css'

import Alert from '@/components/Alert.vue'
import { AlertTypes } from './types/AlertTypes'

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
  private showDetails = false

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

  private viewState = {
    latitude: 0,
    longitude: 0,
    zoom: 2,
    pitch: 0,
    bearing: 0,
    minZoom: 0
  }

  private message: string
  private alertType: string

  constructor () {
    super()
    this.searchID = ''
    this.message = ''
    this.alertType = AlertTypes.ERROR
  }

  private mounted () {
    this.searchID = this.$route.params.pathMatch || ''
    if (this.searchID !== '') {
      this.fetchSessionDetails()
      this.detailsLoop = setInterval(() => {
        this.fetchSessionDetails()
      }, 10000)
    }
  }

  private beforeDestroy () {
    clearInterval(this.detailsLoop)
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

  private fetchSessionDetails () {
    this.$apiService.fetchSessionDetails({ session_id: this.searchID })
      .then((response: any) => {
        this.meta = response.meta
        this.slices = response.slices

        this.meta.connection = this.meta.connection === 'wifi' ? 'Wifi' : this.meta.connection.charAt(0).toUpperCase() + this.meta.connection.slice(1)

        if (!this.showDetails) {
          this.showDetails = true
        }

        setTimeout(() => {
          this.generateCharts()
          const NNCOLOR = [0, 109, 44]
          const DIRECTCOLOR = [49, 130, 189]

          const cellSize = 10
          const aggregation = 'MEAN'
          const gpuAggregation = navigator.appVersion.indexOf('Win') === -1

          this.viewState.latitude = this.meta.location.latitude
          this.viewState.longitude = this.meta.location.longitude

          if (!this.mapInstance) {
            this.mapInstance = new mapboxgl.Map({
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
          const sessionLocationLayer = new ScreenGridLayer({
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
            this.deckGlInstance = new Deck({
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
        })
      })
      .catch((error: any) => {
        if (this.detailsLoop) {
          clearInterval(this.detailsLoop)
        }
        if (this.slices.length === 0) {
          this.message = 'Failed to fetch session details'
          console.log(`Something went wrong fetching sessions details for: ${this.searchID}`)
          console.log(error)
        }
      })
  }

  private generateCharts () {
    const latencyData: Array<Array<number>> = [
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

    let lastEntryNN = false
    let countNN = 0
    let directOnly = true

    this.slices.map((slice: any) => {
      const timestamp = new Date(slice.timestamp).getTime() / 1000
      const onNN = slice.on_network_next

      if (directOnly && onNN) {
        directOnly = false
      }

      let nextRTT = parseFloat(slice.next.rtt)
      const directRTT = parseFloat(slice.direct.rtt)

      let nextJitter = parseFloat(slice.next.jitter)
      const directJitter = parseFloat(slice.direct.jitter)

      let nextPL = parseFloat(slice.next.packet_loss)
      const directPL = parseFloat(slice.direct.packet_loss)

      if (lastEntryNN && !onNN) {
        countNN = 0
      }

      if (onNN && countNN < 3) {
        nextRTT = nextRTT >= directRTT ? directRTT : nextRTT
        nextJitter = nextJitter >= directJitter ? directJitter : nextJitter
        nextPL = 0
        countNN++
      }

      // Latency
      let next = (slice.is_multipath && nextRTT >= directRTT) ? directRTT : nextRTT
      let direct = directRTT
      latencyData[0].push(timestamp)
      latencyData[1].push(next)
      latencyData[2].push(direct)

      // Jitter
      next = (slice.is_multipath && nextJitter >= directJitter) ? directJitter : nextJitter
      direct = directJitter
      jitterData[0].push(timestamp)
      jitterData[1].push(next)
      jitterData[2].push(direct)

      // Packetloss
      next = (slice.is_multipath && nextPL >= directPL) ? directPL : nextPL
      direct = directPL
      packetLossData[0].push(timestamp)
      packetLossData[1].push(next)
      packetLossData[2].push(direct)

      // Bandwidth
      bandwidthData[0].push(timestamp)
      bandwidthData[1].push(slice.envelope.up)
      bandwidthData[2].push(slice.envelope.down)

      lastEntryNN = onNN
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
        ms: {
          from: 'y',
          auto: false,
          range: (self: uPlot, min: number, max: number): uPlot.MinMax => [
            0,
            max
          ]
        }
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
        kbps: {
          from: 'y',
          auto: false,
          range: (self: uPlot, min: number, max: number): uPlot.MinMax => [
            0,
            max
          ]
        }
      },
      series: [
        {},
        {
          stroke: 'blue',
          fill: 'rgba(0,0,255,0.1)',
          label: 'Up'
        },
        {
          stroke: 'orange',
          fill: 'rgba(255,165,0,0.1)',
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
      this.jitterComparisonChart = new uPlot(latencyComparisonOpts, jitterData, jitterChartElement)
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

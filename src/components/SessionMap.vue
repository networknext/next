<template>
  <div v-bind:class="{
         'map-container-no-offset': true,
         'map-container-offset': false,
       }">
    <div id="map"></div>
    <canvas id="deck-canvas"></canvas>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Deck } from '@deck.gl/core'
import { ScreenGridLayer } from '@deck.gl/aggregation-layers'
import mapboxgl from 'mapbox-gl'
import APIService from '@/services/api.service'

/**
 * TODO: Hookup API call
 * TODO: Hookup looping logic
 * TODO: Cleanup template
 */

@Component({
  name: 'SessionMap'
})
export default class SessionMap extends Vue {
  // TODO: Add in types for all of the any declarations
  private accessToken = 'pk.eyJ1Ijoibm5zZWN1cml0eSIsImEiOiJja2FmaXE1Y2cwZGRiMzBub2p3cnE4c3czIn0.3QIueg8fpEy5cBtqRuXMxw'

  private mapInstance: any = null
  private deckGlInstance: any = null

  private mapLoop = -1

  private apiService: APIService

  private viewState = {
    latitude: 0,
    longitude: 0,
    zoom: 2,
    pitch: 0,
    bearing: 0,
    minZoom: 0
  }

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
  }

  private mounted () {
    this.refreshMapSessions()
    this.mapLoop = setInterval(() => {
      this.refreshMapSessions()
    }, 10000)
  }

  private beforeDestroy () {
    clearInterval(this.mapLoop)
  }

  private refreshMapSessions () {
    // creating the map
    this.apiService.call('BuyersService.SessionMap', {})
      .then((response: any) => {
        if (!this.mapInstance) {
          this.mapInstance = new mapboxgl.Map({
            accessToken: this.accessToken,
            style: 'mapbox://styles/mapbox/dark-v10',
            center: [
              0,
              0
            ],
            zoom: 2,
            pitch: 0,
            bearing: 0,
            container: 'map'
          })
        }

        const sessions = response.result.map_points || []
        const onNN = sessions.filter((point: any) => {
          return (point[2] === 1)
        })
        const direct = sessions.filter((point: any) => {
          return (point[2] === 0)
        })

        const cellSize = 10
        const aggregation = 'MEAN'
        const gpuAggregation = navigator.appVersion.indexOf('Win') === -1

        const nnLayer = new ScreenGridLayer({
          id: 'nn-layer',
          data: onNN,
          opacity: 0.8,
          getPosition: (d: Array<number>) => [d[0], d[1]],
          getWeight: () => 1,
          cellSizePixels: cellSize,
          colorRange: [
            [
              40,
              167,
              69
            ]
          ],
          gpuAggregation,
          aggregation
        })

        const directLayer = new ScreenGridLayer({
          id: 'direct-layer',
          data: direct,
          opacity: 0.8,
          getPosition: (d: Array<number>) => [d[0], d[1]],
          getWeight: () => 1,
          cellSizePixels: cellSize,
          colorRange: [
            [
              49,
              130,
              189
            ]
          ],
          gpuAggregation,
          aggregation
        })

        const layers = (onNN.length > 0 || direct.length > 0) ? [directLayer, nnLayer] : []

        if (!this.deckGlInstance) {
          // creating the deck.gl instance
          this.deckGlInstance = new Deck({
            canvas: document.getElementById('deck-canvas'),
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
                minZoom: 2
              })
            },
            layers: layers
          })
        } else {
          this.deckGlInstance.setProps({ layers: [] })
          this.deckGlInstance.setProps({ layers: layers })
        }
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .map-container-offset {
    width: 100%;
    height: calc(-160px + 90vh);
    position: relative;
    overflow: hidden;
  }
  .map-container-no-offset {
    width: 100%;
    height: calc(-160px + 100vh);
    position: relative;
    overflow: hidden;
  }
  #map {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    border: 1px solid rgb(136, 136, 136);
    background-color: rgb(27, 27, 27);
    overflow: hidden;
  }
  #deck-canvas {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
  }
</style>

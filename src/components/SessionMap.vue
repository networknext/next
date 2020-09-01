<template>
  <div v-bind:class="{
    'map-container-no-offset': !$store.getters.isAnonymousPlus,
    'map-container-offset': $store.getters.isAnonymousPlus,
  }">
    <div class="map" id="map"></div>
    <canvas id="deck-canvas"></canvas>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Deck } from '@deck.gl/core'
import { ScreenGridLayer } from '@deck.gl/aggregation-layers'
import mapboxgl from 'mapbox-gl'

/**
 * This component displays the map that is visible in the map workspace
 *  and has all of the associated logic and api calls
 */

/**
 * TODO: Cleanup component logic
 */

@Component({
  name: 'SessionMap'
})
export default class SessionMap extends Vue {
  private deckGlInstance: any
  private mapInstance: any
  private mapLoop: number
  private viewState: any
  private unwatch: any

  constructor () {
    super()
    this.viewState = {
      latitude: 0,
      longitude: 0,
      zoom: 2,
      pitch: 0,
      bearing: 0,
      minZoom: 2,
      maxZoom: 16
    }
    this.mapLoop = -1
  }

  private mounted () {
    this.restartLoop()
    this.unwatch = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        clearInterval(this.mapLoop)
        this.restartLoop()
      }
    )
  }

  private beforeDestroy () {
    clearInterval(this.mapLoop)
    this.unwatch()
  }

  private fetchMapSessions () {
    // TODO: Figure out how to get rid of this. this.$apiService should be possible...
    // HACK: This is a hack to get tests to work properly
    (this as any).$apiService
      .fetchMapSessions({
        buyer_id: this.$store.getters.currentFilter.buyerID || ''
      })
      .then((response: any) => {
        if (!this.mapInstance) {
          this.mapInstance = new mapboxgl.Map({
            accessToken: process.env.VUE_APP_MAPBOX_TOKEN,
            style: 'mapbox://styles/mapbox/dark-v10',
            center: [0, 0],
            zoom: 2,
            pitch: 0,
            bearing: 0,
            container: 'map'
          })
        }

        const sessions = response.map_points || []
        const onNN = sessions.filter((point: any) => {
          return point[2] === 1
        })
        const direct = sessions.filter((point: any) => {
          return point[2] === 0
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
          colorRange: [[40, 167, 69]],
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
          colorRange: [[49, 130, 189]],
          gpuAggregation,
          aggregation
        })

        const layers =
          onNN.length > 0 || direct.length > 0 ? [directLayer, nnLayer] : []

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
                minZoom: 2,
                maxZoom: 16
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

  private restartLoop () {
    this.fetchMapSessions()
    this.mapLoop = setInterval(() => {
      this.fetchMapSessions()
    }, 10000)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
.map-container-offset {
  width: 100%;
  height: calc(-160px + 95vh);
  position: relative;
  overflow: hidden;
  max-height: 1000px;
}
.map-container-no-offset {
  width: 100%;
  height: calc(-160px + 100vh);
  position: relative;
  overflow: hidden;
  max-height: 1000px;
}
.map {
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
}
#deckgl-overlay {
  width: 100%;
  height: 100%;
}
</style>

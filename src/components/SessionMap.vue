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
  private mapLoop: any
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
    this.$apiService
      .fetchMapSessions({
        company_code: this.$store.getters.currentFilter.companyCode || ''
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
          // this.mapInstance.setRenderWorldCopies(status === 'false')
        }

        const sessions = response.map_points || []
        let onNN = []
        let direct = []

        if (this.$store.getters.isAnonymous || this.$store.getters.isAnonymousPlus || this.$store.getters.currentFilter.companyCode === '') {
          onNN = sessions
        } else {
          onNN = sessions.filter((point: any) => {
            return (point[2] === true)
          })
          direct = sessions.filter((point: any) => {
            return (point[2] === false)
          })
        }

        console.log(onNN[0])
        console.log(onNN[1])
        console.log(onNN[2])
        console.log(onNN[3])
        console.log(onNN[4])
        console.log(onNN[5])

        const cellSize = 10
        const aggregation = 'MEAN'
        let nnLayer = null
        let directLayer = null
        let gpuAggregation = navigator.appVersion.indexOf('Win') === -1
        gpuAggregation = gpuAggregation ? navigator.appVersion.indexOf('Macintosh') === -1 : false

        const MAX_SESSIONS = 300000

        console.log('nn received')
        console.log(onNN.length)
        console.log('expected number of nn layers - 1')
        console.log(Math.floor(onNN.length / MAX_SESSIONS))
        const nnLayerNum = Math.floor(onNN.length / MAX_SESSIONS) + 1
        const onNNLayers = []

        console.log('numNNLayers')
        console.log(nnLayerNum)

        for (let i = 0; i < nnLayerNum - 1; i++) {
          console.log('In For Loop NN')
          const currentSlice = onNN.splice(0, MAX_SESSIONS)
          onNN = onNN.splice(-MAX_SESSIONS)
          nnLayer = new ScreenGridLayer({
            id: 'nn-layer-' + i,
            data: currentSlice,
            opacity: 0.8,
            getPosition: (d: Array<number>) => [d[0], d[1]],
            getWeight: () => 1,
            cellSizePixels: cellSize,
            colorRange: [[40, 167, 69]],
            gpuAggregation,
            aggregation
          })
          onNNLayers.push(nnLayer)
        }
        if (onNN.length > 0) {
          nnLayer = new ScreenGridLayer({
            id: 'nn-layer-' + (nnLayerNum - 1),
            data: onNN,
            opacity: 0.8,
            getPosition: (d: Array<number>) => [d[0], d[1]],
            getWeight: () => 1,
            cellSizePixels: cellSize,
            colorRange: [[40, 167, 69]],
            gpuAggregation,
            aggregation
          })
          onNNLayers.push(nnLayer)
        }

        console.log('direct received')
        console.log(direct.length)
        console.log('expected number of direct layers - 1')
        console.log(Math.floor(direct.length / MAX_SESSIONS))

        const directLayerNum = Math.floor(direct.length / MAX_SESSIONS) + 1
        const directLayers = []

        console.log('numDirectLayers')
        console.log(directLayerNum)

        for (let i = 0; i < directLayerNum - 1; i++) {
          const currentSlice = direct.splice(0, MAX_SESSIONS)
          direct = direct.splice(-MAX_SESSIONS)
          directLayer = new ScreenGridLayer({
            id: 'direct-layer-' + i,
            data: currentSlice,
            opacity: 0.8,
            getPosition: (d: Array<number>) => [d[0], d[1]],
            getWeight: () => 1,
            cellSizePixels: cellSize,
            colorRange: [[40, 167, 69]],
            gpuAggregation,
            aggregation
          })
          directLayers.push(directLayer)
        }
        if (direct.length > 0) {
          directLayer = new ScreenGridLayer({
            id: 'direct-layer-' + (directLayerNum - 1),
            data: direct,
            opacity: 0.8,
            getPosition: (d: Array<number>) => [d[0], d[1]],
            getWeight: () => 1,
            cellSizePixels: cellSize,
            colorRange: [[40, 167, 69]],
            gpuAggregation,
            aggregation
          })
          directLayers.push(directLayer)
        }

        let layers: any = []
        if (directLayers.length > 0) {
          layers = layers.concat(directLayers)
        }
        if (onNN.length > 0) {
          layers = layers.concat(onNNLayers)
        }

        console.log('onNNLayers')
        console.log(onNNLayers.length)

        console.log('all layers')
        console.log(layers.length)
        console.log(layers)

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
    if (this.mapLoop) {
      clearInterval(this.mapLoop)
    }
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

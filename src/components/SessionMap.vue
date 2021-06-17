<template>
  <div class="map-container-no-offset">
    <div class="map" id="map"></div>
    <canvas style="cursor: grab;" id="deck-canvas" data-intercom="map"></canvas>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import CustomScreenGridLayer from './CustomScreenGridLayer'

/* import data1 from '../../test_data/ghost-army-map-points-1.json'
import data2 from '../../test_data/ghost-army-map-points-2.json'
import data3 from '../../test_data/ghost-army-map-points-3.json'
import data4 from '../../test_data/ghost-army-map-points-4.json'
import data5 from '../../test_data/ghost-army-map-points-5.json'
import data6 from '../../test_data/ghost-army-map-points-6.json'
import data7 from '../../test_data/ghost-army-map-points-7.json'
import data8 from '../../test_data/ghost-army-map-points-8.json'
import data9 from '../../test_data/ghost-army-map-points-9.json'
import data10 from '../../test_data/ghost-army-map-points-10.json'
import data11 from '../../test_data/ghost-army-map-points-11.json'
import data12 from '../../test_data/ghost-army-map-points-12.json' */

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
  get offsetMap () {
    return this.$store.getters.isAnonymousPlus
  }

  private deckGlInstance: any
  private mapInstance: any
  private mapLoop: any
  private viewState: any
  private unwatchFilter: any

  // private sessions: Array<any>

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

    // Use this to test using the canned json files
    /* this.sessions = (data1 as any).result.map_points
    this.sessions = this.sessions.concat((data2 as any).result.map_points)
    this.sessions = this.sessions.concat((data3 as any).result.map_points)
    this.sessions = this.sessions.concat((data4 as any).result.map_points)
    this.sessions = this.sessions.concat((data5 as any).result.map_points)
    this.sessions = this.sessions.concat((data6 as any).result.map_points)
    this.sessions = this.sessions.concat((data7 as any).result.map_points)
    this.sessions = this.sessions.concat((data8 as any).result.map_points)
    this.sessions = this.sessions.concat((data9 as any).result.map_points)
    this.sessions = this.sessions.concat((data10 as any).result.map_points)
    this.sessions = this.sessions.concat((data11 as any).result.map_points)
    this.sessions = this.sessions.concat((data12 as any).result.map_points) */
  }

  private mounted () {
    this.restartLoop()
    this.unwatchFilter = this.$store.watch(
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
    this.unwatchFilter()
  }

  private fetchMapSessions () {
    this.$apiService
      .fetchMapSessions({
        company_code: this.$store.getters.currentFilter.companyCode || ''
      })
      .then((response: any) => {
        // check if mapbox exists - primarily for tests
        if (!(window as any).mapboxgl) {
          return
        }

        if (!this.mapInstance) {
          this.mapInstance = new (window as any).mapboxgl.Map({
            accessToken: process.env.VUE_APP_MAPBOX_TOKEN,
            style: 'mapbox://styles/mapbox/dark-v10',
            center: [0, 0],
            zoom: 2,
            pitch: 0,
            bearing: 0,
            container: 'map',
            minZoom: 1.9
          })
          // this.mapInstance.setRenderWorldCopies(status === 'false')
        }

        const sessions = response.map_points || []
        let onNN: Array<any> = []
        const direct: Array<any> = []

        if (this.$store.getters.isAnonymous || this.$store.getters.isAnonymousPlus || this.$store.getters.currentFilter.companyCode === '') {
          onNN = sessions
        } else {
          sessions.forEach((session: any) => {
            if (session[2] === true) {
              onNN.push(session)
            } else {
              direct.push(session)
            }
          })
        }

        const cellSize = 10
        const aggregation = 'MEAN'
        const gpuAggregation = navigator.appVersion.indexOf('Win') === -1

        const layers: Array<any> = []

        if (direct.length > 0) {
          const directLayer = new CustomScreenGridLayer({
            id: 'direct-layer',
            data: direct,
            opacity: 0.8,
            getPosition: (d: Array<number>) => [d[0], d[1]],
            pickable: true,
            getWeight: () => 1,
            cellSizePixels: cellSize,
            colorRange: [[49, 130, 189]],
            gpuAggregation,
            aggregation,
            onClick: this.mapPointClickHandler
          })
          layers.push(directLayer)
        }

        // const MAX_SESSIONS = this.sessions.length

        if (onNN.length > 0) {
        /* let slice = []
          for (let i = 0; i < this.sessions.length; i++) {
            slice.push(this.sessions[i])
            if ((i !== 0 && i % MAX_SESSIONS === 0) || i + 1 === this.sessions.length) {
              const nnLayer = new (window as any).deck.ScreenGridLayer({
                id: 'nn-layer-' + i,
                data: slice,
                opacity: 0.8,
                getPosition: (d: Array<number>) => [d[0], d[1]],
                getWeight: () => 1,
                cellSizePixels: cellSize,
                colorRange: [[40, 167, 69]],
                gpuAggregation,
                aggregation
              })
              layers.push(nnLayer)
              slice = []
            }
          } */
          const nnLayer = new CustomScreenGridLayer({
            id: 'nn-layer',
            data: onNN,
            getPosition: (d: Array<number>) => [d[0], d[1]],
            pickable: true,
            cellSizePixels: cellSize,
            colorRange: [[40, 167, 69]],
            aggregation,
            gpuAggregation,
            onClick: this.mapPointClickHandler
          })
          layers.push(nnLayer)
        }

        if (!this.deckGlInstance) {
          // creating the deck.gl instance
          this.deckGlInstance = new (window as any).deck.Deck({
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
            layers: layers,
            getCursor: ({ isHovering, isDragging }: any) => {
              if (isHovering) {
                return 'pointer'
              }
              if (isDragging) {
                return 'grabbing'
              }
              return 'grab'
            }
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

  private mapPointClickHandler (info: any) {
    const points: any = info.object.points || []
    if (points.length === 0) {
      this.$root.$emit('failedMapPointLookup')
      return
    }
    if (points.length === 1 && points[0].source[3]) {
      this.$router.push(`/session-tool/${points[0].source[3]}`)
      return
    }
    this.$root.$emit('showModal', points)
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

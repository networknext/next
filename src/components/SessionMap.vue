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
// import data1 from '../../map_sessions.json'
// import data2 from '../../map_sessions_2.json'
// import data3 from '../../map_sessions_3.json'
// import data4 from '../../map_sessions_4.json'
// import data5 from '../../map_sessions_5.json'
// import data6 from '../../map_sessions_6.json'

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
  private unwatchFilter: any
  private unwatchSignUp: any

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
  /*     this.sessions = (data1 as any).map_points
    this.sessions = this.sessions.concat((data2 as any).map_points)
    this.sessions = this.sessions.concat((data3 as any).map_points)
    this.sessions = this.sessions.concat((data4 as any).map_points)
    this.sessions = this.sessions.concat((data5 as any).map_points)
    this.sessions = this.sessions.concat((data6 as any).map_points) */
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
    this.unwatchSignUp = this.$store.watch(
      (state: any, getters: any) => {
        return getters.isSignUp
      },
      (isSignUp: boolean) => {
        if (isSignUp) {
          this.$apiService.addMailChimpContact({ email: this.$store.getters.userProfile.email }).then(() => {
            this.$store.commit('UPDATE_IS_SIGNUP', false)
          }).catch((error: Error) => {
            console.log('Failed to add new sign up to mail list')
            console.log(error)
          })
        }
      }
    )
  }

  private beforeDestroy () {
    clearInterval(this.mapLoop)
    this.unwatchFilter()
    this.unwatchSignUp()
  }

  private fetchMapSessions () {
    this.$apiService
      .fetchMapSessions({
        company_code: this.$store.getters.currentFilter.companyCode || ''
      })
      .then((response: any) => {
        if (!this.mapInstance) {
          this.mapInstance = new (window as any).mapboxgl.Map({
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

        let layers: any = []

        if (direct.length > 0) {
          const directLayer = new (window as any).deck.ScreenGridLayer({
            id: 'direct-layer',
            data: direct,
            opacity: 0.8,
            getPosition: (d: Array<number>) => [d[0], d[1]],
            getWeight: () => 1,
            cellSizePixels: cellSize,
            colorRange: [[49, 130, 189]],
            gpuAggregation,
            aggregation,
            coordinateSystem: 1
          })
          layers = layers.push(directLayer)
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
          const nnLayer = new (window as any).deck.ScreenGridLayer({
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

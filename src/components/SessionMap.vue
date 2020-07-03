<template>
  <div id='map-container'
    v-bind:class="{
      'map-container-no-offset': true,
      'map-container-offset': false,
    }"
  ></div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Deck } from '@deck.gl/core'

@Component({
  name: 'SessionMap'
})
export default class SessionMap extends Vue {
  accessToken = 'pk.eyJ1Ijoibm5zZWN1cml0eSIsImEiOiJja2FmaXE1Y2cwZGRiMzBub2p3cnE4c3czIn0.3QIueg8fpEy5cBtqRuXMxw'

  mapInstance: any = null
  deckGlInstance: any = null

  constructor () {
    super()
    this.refreshMapSessions()
    /* setInterval(() => {
      this.refreshMapSessions()
    }, 10000) */
  }

  private refreshMapSessions () {
    // creating the map
    this.mapInstance = new Deck({
      mapboxApiAccessToken: this.accessToken,
      mapStyle: 'mapbox://styles/mapbox/dark-v10',
      initialViewState: {
        zoom: 2,
        longitude: 0, // 'Center' of the world map
        latitude: 0,
        minZoom: 2,
        bearing: 0,
        pitch: 0
      },
      container: '#map-container',
      controller: true,
      layers: []
    })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .map-container-offset {
    width: 100%;
    height: calc(-160px + 90vh);
    border: 1px solid rgb(136, 136, 136);
    background-color: rgb(27, 27, 27);
  }
  .map-container-no-offset {
    width: 100%;
    height: calc(-160px + 100vh);
    border: 1px solid rgb(136, 136, 136);
    background-color: rgb(27, 27, 27);
  }
</style>

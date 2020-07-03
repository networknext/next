<template>
    <div id="map-container"
         class="map-container-no-offset"
         v-bind:class="{
            'map-container-no-offset': true,
            'map-container-offset': false,
         }"
    ></div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component({
  name: 'SessionMap'
})
export default class SessionMap extends Vue {
  constructor () {
    super()
    this.refreshMapSessions()
    setInterval(() => {
      this.refreshMapSessions()
    }, 10000)
  }

  private refreshMapSessions () {
    Vue.prototype.$apiService.call('BuyersService.SessionMap', {}).then((response: any) => {
      console.log(response)
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

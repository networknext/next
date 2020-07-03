<template>
  <div id="map-container">
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
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
</style>

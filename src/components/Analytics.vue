<template>
  <div class="card-body" id="analytics-page">
    <iframe
      v-if="dashboardURL !== ''"
      v-bind:src="dashboardURL"
      width="1000"
      height="2000"
      frameborder="0">
    </iframe>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class Analytics extends Vue {
  private dashboardURL: string

  constructor () {
    super()
    this.dashboardURL = ''
  }

  private mounted () {
    this.$apiService.fetchLookerURL().then((response: any) => {
      console.log(response)
      this.dashboardURL = response.url
    })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

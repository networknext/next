<template>
  <h1 class="count-header">
    {{ $store.getters.currentPage[0].toUpperCase() + $store.getters.currentPage.slice(1) }}&nbsp;
    <span class="badge badge-dark">
      {{ this.totalSessions }} Total Sessions
    </span>&nbsp;
    <span class="badge badge-success">
      {{ this.totalSessionsReply.onNN }} on Network Next
    </span>
  </h1>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import APIService from '../services/api.service'

interface TotalSessionsReply {
  direct: number;
  onNN: number;
}

@Component
export default class SessionCounts extends Vue {
  private totalSessionsReply: TotalSessionsReply
  private apiService: APIService
  private countLoop = -1

  // TODO: These values should probably go in a store

  get totalSessions () {
    return this.totalSessionsReply.direct + this.totalSessionsReply.onNN
  }

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
    this.totalSessionsReply = {
      direct: 0,
      onNN: 0
    }
  }

  private mounted () {
    this.fetchSessionCounts()
    this.countLoop = setInterval(() => {
      this.fetchSessionCounts()
    }, 1000)
  }

  private beforeDestroy () {
    clearInterval(this.countLoop)
  }

  private fetchSessionCounts () {
    this.apiService.call('BuyersService.TotalSessions', {})
      .then((response: any) => {
        this.totalSessionsReply.direct = response.result.direct
        this.totalSessionsReply.onNN = response.result.next
      })
      .catch((error: any) => {
        console.log(error)
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .count-header {
    font-size: 2rem;
  }
</style>

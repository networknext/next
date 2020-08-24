<template>
  <h1 class="count-header" v-if="showCount" data-test="currentPage">
    {{ $store.getters.currentPage[0].toUpperCase() + $store.getters.currentPage.slice(1) }}
    <span class="badge badge-dark" data-test="totalSessions">
      {{ this.totalSessions }} Total Sessions
    </span>&nbsp;
    <span class="badge badge-success" data-test="nnSessions">
      {{ this.totalSessionsReply.onNN }} on Network Next
    </span>
  </h1>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import APIService from '../services/api.service'

/**
 * This component displays the total session counts and has all of the associated logic and api calls
 */

/**
 * TODO: Add filter bar back in here, potentially in its own component if it is worth while?
 * TODO: Figure out how to turn this into a class with functions that help control the count and refresh loop
 *  This would help with the filter bar...
 *  Similar idea with the alert component
 */

interface TotalSessionsReply {
  direct: number;
  onNN: number;
}

@Component
export default class SessionCounts extends Vue {
  private totalSessionsReply: TotalSessionsReply
  private apiService: APIService
  private showCount: boolean
  private countLoop: number

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
    this.countLoop = -1
    this.showCount = false
  }

  private mounted () {
    this.fetchSessionCounts()
    this.countLoop = setInterval(() => {
      this.fetchSessionCounts()
    }, 10000000000000)
  }

  private beforeDestroy () {
    clearInterval(this.countLoop)
  }

  private fetchSessionCounts () {
    this.apiService.fetchTotalSessionCounts({})
      .then((response: any) => {
        this.totalSessionsReply.direct = response.direct
        this.totalSessionsReply.onNN = response.next
        if (!this.showCount) {
          this.showCount = true
        }
      })
      .catch((error: Error) => {
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

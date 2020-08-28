<template>
  <div
    class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"
  >
    <h1 class="count-header" v-if="showCount" data-test="currentPage">
      {{ $store.getters.currentPage[0].toUpperCase() + $store.getters.currentPage.slice(1) }}&nbsp;
      <span
        class="badge badge-dark"
        data-test="totalSessions"
      >{{ this.totalSessions }} Total Sessions</span>&nbsp;
      <span
        class="badge badge-success"
        data-test="nnSessions"
      >{{ this.totalSessionsReply.onNN }} on Network Next</span>
    </h1>
    <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1">
      <div class="mr-auto"></div>
      <div class="px-2" v-if="$store.getters.isBuyer || $store.getters.isAdmin">
        <select class="form-control" v-on:change="updateFilter($event.target.value)">
          <option
            :value="getBuyerId()"
            v-if="!$store.getters.isAdmin && $store.getters.isBuyer"
            :selected="getBuyerId() == $store.getters.currentFilter"
          >{{ getBuyerName() }}</option>
          <option :value="''" :selected="'' == $store.getters.currentFilter">All</option>
          <option
            :value="buyer.id"
            v-for="buyer in allBuyers"
            v-bind:key="buyer.id"
            :selected="buyer.id == $store.getters.currentFilter"
          >{{ buyer.name }}</option>
        </select>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

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
  private totalSessionsReply: TotalSessionsReply;
  private showCount: boolean;
  private countLoop: number;
  private vueInstance: any;

  get totalSessions () {
    return this.totalSessionsReply.direct + this.totalSessionsReply.onNN
  }

  get allBuyers () {
    return this.$store.getters.allBuyers.filter((buyer: any) => {
      return buyer.is_live || this.$store.getters.isAdmin
    })
  }

  constructor () {
    super()
    this.totalSessionsReply = {
      direct: 0,
      onNN: 0
    }
    this.countLoop = -1
    this.showCount = false
    this.vueInstance = Vue
  }

  private mounted () {
    this.restartLoop()
  }

  private beforeDestroy () {
    clearInterval(this.countLoop)
  }

  private fetchSessionCounts () {
    this.vueInstance
      .fetchTotalSessionCounts({
        buyer_id: this.$store.getters.currentFilter.buyerID || ''
      })
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

  private getBuyerId () {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].id === this.$store.getters.userProfile.buyerID) {
        return allBuyers[i].id
      }
    }
    return 'Private'
  }

  private getBuyerName () {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].id === this.$store.getters.userProfile.buyerID) {
        return allBuyers[i].name
      }
    }
    return 'Private'
  }

  private updateFilter (buyerID: string) {
    this.$store.commit('UPDATE_CURRENT_FILTER', { buyerID: buyerID })
    this.restartLoop()
  }

  private restartLoop () {
    this.fetchSessionCounts()
    this.countLoop = setInterval(() => {
      this.fetchSessionCounts()
    }, 1000)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
.count-header {
  font-size: 2rem;
}
</style>

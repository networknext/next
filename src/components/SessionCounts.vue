<template>
  <div
    class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"
  >
    <h1 class="count-header" v-if="showCount" data-test="currentPage">
      {{ $store.getters.currentPage[0].toUpperCase() + $store.getters.currentPage.slice(1) }}&nbsp;
      <span
        class="badge badge-dark"
        data-test="totalSessions"
      >{{ this.totalSessions.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',') }} Total Sessions</span>&nbsp;
      <span
        class="badge badge-success"
        data-test="nnSessions"
      >{{ this.totalSessionsReply.onNN.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',') }} on Network Next</span>
    </h1>
    <div class="mb-2 mb-md-0 flex-grow-1 align-items-center pl-4 pr-4" v-if="$store.getters.isAnonymousPlus">
      <Alert :message="`Please confirm your email address: ${$store.getters.userProfile.email}`" :alertType="AlertType.INFO" ref="verifyAlert">
        <a href="#" @click="$refs.verifyAlert.resendVerificationEmail()">
          Resend email
        </a>
      </Alert>
    </div>
    <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1" v-if="!$store.getters.isAnonymousPlus">
      <div class="mr-auto"></div>
      <div class="px-2" v-if="$store.getters.isBuyer || $store.getters.isAdmin">
        <select class="form-control" v-on:change="updateFilter($event.target.value)">
          <option v-for="option in filterOptions" :key="option.value" v-bind:value="option.value" v-bind:selected="$store.getters.currentFilter.companyCode === option.value">
            {{ option.name }}
          </option>
        </select>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { AlertType } from './types/AlertTypes'
import Alert from '@/components/Alert.vue'

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

@Component({
  components: {
    Alert
  }
})
export default class SessionCounts extends Vue {
  private totalSessionsReply: TotalSessionsReply
  private showCount: boolean
  private countLoop: any
  private AlertType: any

  private filterOptions: Array<any>

  constructor () {
    super()
    this.totalSessionsReply = {
      direct: 0,
      onNN: 0
    }
    this.showCount = false
    this.AlertType = AlertType
  }

  private mounted () {
    this.filterOptions.push({
      name: 'All',
      value: ''
    })

    this.$store.getters.allBuyers.forEach((buyer: any) => {
      if (!this.$store.getters.isAdmin || (this.$store.getters.isAdmin && buyer.is_live)) {
        this.filterOptions.push({
          name: buyer.company_name,
          value: buyer.company_code
        })
      }
    })

    this.restartLoop()
  }

  private beforeDestroy () {
    clearInterval(this.countLoop)
  }

  private fetchSessionCounts () {
    this.$apiService.fetchTotalSessionCounts({ company_code: this.$store.getters.currentFilter.companyCode })
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

  private getBuyerCode () {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].company_code === this.$store.getters.userProfile.companyCode) {
        return allBuyers[i].company_code
      }
    }
    return 'Private'
  }

  private getBuyerName () {
    const allBuyers = this.$store.getters.allBuyers
    let i = 0
    for (i; i < allBuyers.length; i++) {
      if (allBuyers[i].company_code === this.$store.getters.userProfile.companyCode) {
        return allBuyers[i].company_name
      }
    }
    return 'Private'
  }

  private updateFilter (companyCode: string) {
    this.$store.commit('UPDATE_CURRENT_FILTER', { companyCode: companyCode })
    this.restartLoop()
  }

  private restartLoop () {
    if (this.countLoop) {
      clearInterval(this.countLoop)
    }
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

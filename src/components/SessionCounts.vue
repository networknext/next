<template>
  <div
    class="d-flex justify-content-between flex-wrap flex-md-nowrap align-items-center pt-3 pb-2 mb-3 border-bottom"
  >
    <h1 class="count-header" v-if="showCount" data-test="currentPage">
      {{ $store.getters.currentPage[0].toUpperCase() + $store.getters.currentPage.slice(1) }}&nbsp;
      <span
        class="badge badge-dark"
        data-test="totalSessions"
      >{{ totalSessions.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',') }} Total Sessions</span>&nbsp;
      <span
        class="badge badge-success"
        data-test="nnSessions"
      >{{ totalSessionsReply.onNN.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',') }} on Network Next</span>
    </h1>
    <div class="mb-2 mb-md-0 flex-grow-1 align-items-center pl-4 pr-4">
      <Alert ref="sessionCountAlert">
        <a href="#" @click="$refs.sessionCountAlert.resendVerificationEmail()">
          Resend email
        </a>
      </Alert>
    </div>
    <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1" v-if="!$store.getters.isAnonymousPlus" style="max-width: 300px;">
      <div class="mr-auto"></div>
      <BuyerFilter />
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { AlertType } from './types/AlertTypes'
import Alert from '@/components/Alert.vue'
import BuyerFilter from '@/components/BuyerFilter.vue'

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
    Alert,
    BuyerFilter
  }
})
export default class SessionCounts extends Vue {
  get totalSessions () {
    return this.totalSessionsReply.direct + this.totalSessionsReply.onNN
  }

  // Register the alert component to access its set methods
  $refs!: {
    sessionCountAlert: Alert;
  }

  private totalSessionsReply: TotalSessionsReply
  private showCount: boolean
  private countLoop: any
  private alertToggle: boolean

  private unwatchFilter: any

  constructor () {
    super()
    this.totalSessionsReply = {
      direct: 0,
      onNN: 0
    }
    this.showCount = false
    this.alertToggle = false
  }

  private mounted () {
    this.restartLoop()
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        clearInterval(this.countLoop)
        this.restartLoop()
      }
    )
    if (this.$store.getters.isAnonymousPlus) {
      this.$refs.sessionCountAlert.setMessage(`Please confirm your email address: ${this.$store.getters.userProfile.email}`)
      this.$refs.sessionCountAlert.setAlertType(AlertType.INFO)
    }
    this.restartLoop()

    this.$root.$on('failedMapPointLookup', this.failedMapPointLookupCallback)
  }

  private beforeDestroy () {
    clearInterval(this.countLoop)
    this.unwatchFilter()
    this.$root.$off('failedMapPointLookup')
  }

  private fetchSessionCounts () {
    this.$apiService.fetchTotalSessionCounts({ company_code: this.$store.getters.currentFilter.companyCode })
      .then((response: any) => {
        this.totalSessionsReply.direct = response.direct
        this.totalSessionsReply.onNN = response.next
      })
      .catch((error: Error) => {
        this.totalSessionsReply.direct = 0
        this.totalSessionsReply.onNN = 0
        console.log('Something went wrong fetching session counts')
        console.log(error)
      })
      .finally(() => {
        if (!this.showCount) {
          this.showCount = true
        }
      })
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

  private failedMapPointLookupCallback () {
    if (!this.alertToggle) {
      this.alertToggle = true
      this.$refs.sessionCountAlert.toggleSlots(false)
      this.$refs.sessionCountAlert.setMessage('Map point lookup was unsuccessful. Please zoom in closer and try again')
      this.$refs.sessionCountAlert.setAlertType(AlertType.WARNING)
      setTimeout(() => {
        if (this.$store.getters.isAnonymousPlus) {
          this.$refs.sessionCountAlert.setMessage(`Please confirm your email address: ${this.$store.getters.userProfile.email}`)
          this.$refs.sessionCountAlert.setAlertType(AlertType.INFO)
        } else {
          this.$refs.sessionCountAlert.resetAlert()
        }
        this.$refs.sessionCountAlert.toggleSlots(true)
        this.alertToggle = false
      }, 10000)
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
.count-header {
  font-size: 2rem;
}
</style>

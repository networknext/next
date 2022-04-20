<template>
  <div
    class="d-flex justify-content-between align-items-center pt-3 pb-2 mb-3 border-bottom"
  >
    <h1 class="count-header" v-if="showCount" data-test="currentPage" style="max-width: 80%;">
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
    <div class="mb-2 mb-md-0 align-items-center pl-4 pr-4" style="max-width: 50%">
      <Alert ref="sessionCountAlert">
        <br />
        <a href="#" @click="$refs.sessionCountAlert.resendVerificationEmail()">
          Resend email
        </a>
      </Alert>
    </div>
    <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1" style="max-width: 300px;">
      <div class="mr-auto"></div>
      <BuyerFilter v-if="$store.getters.isBuyer || $store.getters.isAdmin" />
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { AlertType } from './types/AlertTypes'
import Alert from '@/components/Alert.vue'
import BuyerFilter from '@/components/BuyerFilter.vue'
import { EMAIL_CONFIRMATION_MESSAGE, MAX_RETRIES } from '@/components/types/Constants'
import { ErrorTypes } from './types/ErrorTypes'

/**
 * Functions:
 *  Display current sessions counts based on current filter
 *  Display system error statuses
 *  Display map look up failure warning
 *  Display account verification alert
 */

/**
 * TODO:
 *  Move interface to another file with other types
 *  Setup TotalSessionCounts to take this interface to facilitate "unmarshalling"
 *  Clean up logic - primarily different emitters for different alerts
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
  private retryCount: number

  private unwatchDemoMode: any
  private unwatchFilter: any
  private unwatchKillLoops: any

  constructor () {
    super()
    this.totalSessionsReply = {
      direct: 0,
      onNN: 0
    }
    this.showCount = false
    this.alertToggle = false
    this.retryCount = 0
  }

  private mounted () {
    this.unwatchFilter = this.$store.watch(
      (state: any, getters: any) => {
        return getters.currentFilter
      },
      () => {
        clearInterval(this.countLoop)
        this.restartLoop()
      }
    )
    this.unwatchKillLoops = this.$store.watch(
      (state: any, getters: any) => {
        return getters.killLoops
      },
      () => {
        this.showReloadAlert()
      }
    )
    this.unwatchDemoMode = this.$store.watch(
      (state: any, getters: any) => {
        return getters.isDemo
      },
      () => {
        if (this.$store.getters.isDemo) {
          this.$refs.sessionCountAlert.toggleSlots(false)
          this.$refs.sessionCountAlert.setMessage('Demo Mode')
          this.$refs.sessionCountAlert.setAlertType(AlertType.WARNING)
        } else {
          this.$refs.sessionCountAlert.resetAlert()
        }
      }
    )

    if (this.$store.getters.isAnonymousPlus) {
      this.$refs.sessionCountAlert.setMessage(`${EMAIL_CONFIRMATION_MESSAGE} ${this.$store.getters.userProfile.email}`)
      this.$refs.sessionCountAlert.setAlertType(AlertType.INFO)
    }

    if (this.$store.getters.isDemo) {
      this.$refs.sessionCountAlert.toggleSlots(false)
      this.$refs.sessionCountAlert.setMessage('Demo Mode')
      this.$refs.sessionCountAlert.setAlertType(AlertType.WARNING)
    }

    this.$root.$on('failedMapPointLookup', this.failedMapPointLookupCallback)

    // If the network isn't available/working show an alert and skip starting the polling loop
    if (this.$store.getters.killLoops) {
      this.showCount = true
      this.showReloadAlert()
      return
    }

    this.restartLoop()
  }

  private beforeDestroy () {
    clearInterval(this.countLoop)
    this.unwatchDemoMode()
    this.unwatchFilter()
    this.unwatchKillLoops()
    this.$root.$off('failedMapPointLookup')
  }

  private fetchSessionCounts () {
    this.$apiService.fetchTotalSessionCounts({ company_code: this.$store.getters.currentFilter.companyCode })
      .then((response: any) => {
        this.retryCount = 0
        this.totalSessionsReply.direct = response.direct
        this.totalSessionsReply.onNN = response.next
      })
      .catch((error: Error) => {
        console.log('Something went wrong fetching session counts')
        console.log(error)

        this.stopLoop()
        this.retryCount = this.retryCount + 1
        if (this.retryCount < MAX_RETRIES) {
          setTimeout(() => {
            this.restartLoop()
          }, 3000 * this.retryCount)
        } else {
          this.$store.dispatch('toggleKillLoops', true)
        }
      })
      .finally(() => {
        if (!this.showCount) {
          this.showCount = true
        }
      })
  }

  private showReloadAlert () {
    this.stopLoop()
    if (this.$refs.sessionCountAlert.className === AlertType.ERROR) {
      return
    }

    this.$refs.sessionCountAlert.toggleSlots(false)
    this.$refs.sessionCountAlert.setMessage(ErrorTypes.SYSTEM_FAILURE)
    this.$refs.sessionCountAlert.setAlertType(AlertType.ERROR)
  }

  private restartLoop () {
    this.stopLoop()
    this.fetchSessionCounts()
    this.countLoop = setInterval(() => {
      this.fetchSessionCounts()
    }, 1000)
  }

  private stopLoop () {
    if (this.countLoop) {
      clearInterval(this.countLoop)
    }
  }

  private failedMapPointLookupCallback () {
    if (!this.alertToggle) {
      this.alertToggle = true
      this.$refs.sessionCountAlert.toggleSlots(false)
      this.$refs.sessionCountAlert.setMessage(ErrorTypes.FAILED_MAP_POINT_LOOKUP)
      this.$refs.sessionCountAlert.setAlertType(AlertType.WARNING)
      setTimeout(() => {
        if (!this.$refs.sessionCountAlert) {
          return
        }
        if (this.$store.getters.isAnonymousPlus) {
          this.$refs.sessionCountAlert.setMessage(`${EMAIL_CONFIRMATION_MESSAGE} ${this.$store.getters.userProfile.email}`)
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

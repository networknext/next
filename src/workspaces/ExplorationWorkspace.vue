<template>
  <div>
    <div class="
              d-flex
              justify-content-between
              flex-wrap
              flex-md-nowrap
              align-items-center
              pt-3
              pb-2
              mb-3
              border-bottom
            "
    >
      <h1 class="h2">
        Exploration
      </h1>
      <div class="mb-2 mb-md-0 flex-grow-1 align-items-center pl-4 pr-4">
        <Alert ref="verifyAlert"></Alert>
      </div>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1" v-if="$store.getters.currentPage === 'analytics' || $store.getters.currentPage === 'invoice' || $store.getters.currentPage === 'saves' || $store.getters.currentPage === 'usage'" style="max-width: 400px;">
        <div class="mr-auto"></div>
        <BuyerFilter v-if="$store.getters.isAdmin" :includeAll="false" />
      </div>
    </div>
    <div class="card" style="margin-bottom: 250px;">
      <div class="card-header">
        <ul class="nav nav-tabs card-header-tabs">
          <li class="nav-item" v-if="$store.getters.hasBilling">
            <router-link to="/explore/usage" class="nav-link" :class="{ active: $store.getters.currentPage === 'usage' || $store.getters.currentPage === 'invoice'}">Usage</router-link>
          </li>
          <li class="nav-item" v-if="$store.getters.hasAnalytics">
            <router-link to="/explore/analytics" class="nav-link" :class="{ active: $store.getters.currentPage === 'analytics'}">Analytics</router-link>
          </li>
          <li class="nav-item" v-if="false && $store.getters.hasAnalytics">
            <router-link to="/explore/discovery" class="nav-link" :class="{ active: $store.getters.currentPage === 'discovery'}">Discovery</router-link>
          </li>
          <li class="nav-item" v-if="false && $store.getters.isSeller">
            <router-link to="/explore/supply" class="nav-link" :class="{ active: $store.getters.currentPage === 'supply'}">Supplier</router-link>
          </li>
          <li class="nav-item" v-if="$store.getters.isAdmin">
            <router-link to="/explore/saves" class="nav-link" :class="{ active: $store.getters.currentPage === 'saves'}">Saves</router-link>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
import BuyerFilter from '@/components/BuyerFilter.vue'
import DateFilter from '@/components/DateFilter.vue'
import { ANALYTICS_TRIAL_SIGNUP_RESPONSE, RELOAD_MESSAGE } from '@/components/types/Constants'

/**
 * This component holds the workspace elements related to the downloads page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert,
    BuyerFilter,
    DateFilter
  }
})
export default class ExplorationWorkspace extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    verifyAlert: Alert;
  }

  private mounted () {
    // If the network is down, show an error
    if (this.$store.getters.killLoops) {
      this.showErrorAlert()
    }

    this.$root.$on('showAnalyticsTrialResponse', this.showAnalyticsTrialResponse)
  }

  private beforeDestroy () {
    this.$root.$off('showAnalyticsTrialResponse')
  }

  private showErrorAlert () {
    this.$refs.verifyAlert.toggleSlots(false)
    this.$refs.verifyAlert.setMessage(RELOAD_MESSAGE)
    this.$refs.verifyAlert.setAlertType(AlertType.ERROR)
  }

  private showAnalyticsTrialResponse () {
    this.$refs.verifyAlert.toggleSlots(false)
    this.$refs.verifyAlert.setMessage(ANALYTICS_TRIAL_SIGNUP_RESPONSE)
    this.$refs.verifyAlert.setAlertType(AlertType.SUCCESS)
    setTimeout(() => {
      this.$refs.verifyAlert.resetAlert()
    }, 10000)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

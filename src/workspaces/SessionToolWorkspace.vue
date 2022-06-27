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
        Session Tool
      </h1>
    <div class="mb-2 mb-md-0 align-items-center pl-4 pr-4" style="max-width: 50%">
      <Alert ref="verifyAlert">
        <br />
        <a href="#" @click="$refs.sessionCountAlert.resendVerificationEmail()">
          Resend email
        </a>
      </Alert>
    </div>
    <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1" style="max-width: 300px;">
      <div class="mr-auto"></div>
      <BuyerFilter v-if="$store.getters.isBuyer || $store.getters.isAdmin" :liveOnly="true" />
    </div>
    </div>
    <form class="flow-stats-form" @submit.prevent="fetchSessionDetails()">
      <div class="form-group">
        <label for="session-id-input">
            Session ID
        </label>
        <div class="row">
          <div class="col">
            <input class="form-control"
                   type="text"
                   placeholder="Enter a Session ID to view statistics"
                   v-model="searchID"
                   data-test="searchInput"
            >
          </div>
          <!-- TODO: Change this to only hide for anon and anon plus -->
          <div class="col-auto" v-if="$store.getters.isAdmin">
            <LookerDateFilter />
          </div>
          <div class="col-auto">
            <button id="session-tool-button" class="btn btn-primary" type="submit">
              View Stats
            </button>
          </div>
        </div>
      </div>
    </form>
    <Alert ref="inputAlert"/>
    <router-view :key="$route.fullPath"/>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Route, NavigationGuardNext } from 'vue-router'

import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
import { EMAIL_CONFIRMATION_MESSAGE, SESSION_TOOL_ALERT } from '@/components/types/Constants'
import { ErrorTypes } from '@/components/types/ErrorTypes'
import LookerDateFilter from '@/components/LookerDateFilter.vue'
import BuyerFilter from '@/components/BuyerFilter.vue'

/**
 * This component holds the workspace elements related to the session tool in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert,
    BuyerFilter,
    LookerDateFilter
  }
})
export default class SessionToolWorkspace extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    verifyAlert: Alert;
    inputAlert: Alert;
  }

  private searchID: string

  constructor () {
    super()
    this.searchID = ''
  }

  private mounted () {
    if (this.$store.getters.isAnonymousPlus) {
      this.$refs.verifyAlert.setMessage(`${EMAIL_CONFIRMATION_MESSAGE} ${this.$store.getters.userProfile.email}`)
      this.$refs.verifyAlert.setAlertType(AlertType.INFO)
    }
    if (this.$route.path === '/session-tool') {
      this.$refs.inputAlert.setMessage(SESSION_TOOL_ALERT)
      this.$refs.inputAlert.setAlertType(AlertType.INFO)
    }

    // If the network is down, show an error
    if (this.$store.getters.killLoops) {
      this.showErrorAlert()
    }
  }

  private created () {
    this.searchID = this.$route.params.pathMatch || ''
    if (this.searchID !== '') {
      this.fetchSessionDetails()
    }
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    this.searchID = to.params.pathMatch || ''
    if (this.searchID === '') {
      this.$refs.inputAlert.setMessage('Please enter a valid Session ID to view its statistics. It should be a hexadecimal number (with leading zeros), or a decimal number.')
      this.$refs.inputAlert.setAlertType(AlertType.INFO)
    }
    next()
  }

  private fetchSessionDetails () {
    if (this.$refs.inputAlert) {
      this.$refs.inputAlert.resetAlert()
    }
    if (this.searchID === '' && this.$route.path !== '/session-tool') {
      this.$router.push({ path: '/session-tool' })
      return
    }
    if (this.searchID === '' && this.$route.path === '/session-tool') {
      this.$refs.inputAlert.setMessage('Please enter a valid Session ID to view its statistics. It should be a hexadecimal number (with leading zeros), or a decimal number.')
      this.$refs.inputAlert.setAlertType(AlertType.INFO)
      return
    }
    const newRoute = `/session-tool/${this.searchID}`
    if (this.$route.path !== newRoute) {
      this.$router.push({ path: newRoute })
    }
  }

  private showErrorAlert () {
    this.$refs.verifyAlert.toggleSlots(false)
    this.$refs.verifyAlert.setMessage(ErrorTypes.SYSTEM_FAILURE)
    this.$refs.verifyAlert.setAlertType(AlertType.ERROR)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  #session-tool-button {
    border-color: #009FDF;
    background-color: #009fdf;
  }
  #session-tool-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
</style>

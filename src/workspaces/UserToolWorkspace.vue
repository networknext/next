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
      <!-- TODO: Fix this interesting class name -->
      <h1 class="h2">
        User Tool
      </h1>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1" style="max-width: 300px;">
        <div class="mr-auto"></div>
        <BuyerFilter id="buyer-filter" v-if="$store.getters.isAdmin" />
      </div>
    </div>
    <form class="flow-stats-form" @submit.prevent="fetchUserSessions()">
      <div class="form-group">
        <label for="user-id-input">
            User ID
        </label>
        <div class="row">
          <div class="col">
            <input class="form-control"
                    type="text"
                    placeholder="Enter a User ID to view their sessions."
                    v-model="searchID"
            >
          </div>
          <div class="col-auto" v-if="$flagService.isEnabled(FeatureEnum.FEATURE_LOOKER_BIGTABLE_REPLACEMENT)">
            <LookerDateFilter />
          </div>
          <div class="col-auto">
            <button id="user-tool-button" class="btn btn-primary" type="submit">
              View Sessions
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
import UserSessions from '@/components/UserSessions.vue'
import { AlertType } from '@/components/types/AlertTypes'
import Alert from '@/components/Alert.vue'
import { NavigationGuardNext, Route } from 'vue-router'
import { ErrorTypes } from '@/components/types/ErrorTypes'
import { FeatureEnum } from '@/components/types/FeatureTypes'
import LookerDateFilter from '@/components/LookerDateFilter.vue'
import BuyerFilter from '@/components/BuyerFilter.vue'

/**
 * This component holds the workspace elements related to the user tool page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert,
    BuyerFilter,
    LookerDateFilter,
    UserSessions
  }
})
export default class UserToolWorkspace extends Vue {
  // Register the alert component to access its set methods
  $refs!: {
    verifyAlert: Alert;
    inputAlert: Alert;
  }

  private FeatureEnum: any
  private searchID: string

  constructor () {
    super()
    this.searchID = ''
  }

  private mounted () {
    if (this.$route.path === '/user-tool') {
      this.$refs.inputAlert.setMessage('Please enter a User ID to view their sessions.')
      this.$refs.inputAlert.setAlertType(AlertType.INFO)
    }

    // If the network is down, show an error
    if (this.$store.getters.killLoops) {
      this.showErrorAlert()
    }
  }

  private created () {
    this.FeatureEnum = FeatureEnum
    this.searchID = this.$route.params.pathMatch || ''
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    this.searchID = to.params.pathMatch || ''
    if (this.searchID === '') {
      this.$refs.inputAlert.setMessage('Please enter a User ID to view their sessions.')
      this.$refs.inputAlert.setAlertType(AlertType.INFO)
    }
    next()
  }

  private fetchUserSessions () {
    if (this.$refs.inputAlert) {
      this.$refs.inputAlert.resetAlert()
    }
    if (this.searchID === '' && this.$route.path !== '/user-tool') {
      this.$router.push({ path: '/user-tool' })
      return
    }
    if (this.searchID === '' && this.$route.path === '/user-tool') {
      this.$refs.inputAlert.setMessage('Please enter a User ID to view their sessions.')
      this.$refs.inputAlert.setAlertType(AlertType.INFO)
      return
    }
    const newRoute = `/user-tool/${this.searchID}`
    if (this.$route.path !== newRoute) {
      this.$router.push({ path: newRoute })
    }
  }

  private showErrorAlert () {
    this.$refs.inputAlert.toggleSlots(false)
    this.$refs.inputAlert.setMessage(ErrorTypes.SYSTEM_FAILURE)
    this.$refs.inputAlert.setAlertType(AlertType.ERROR)
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  #user-tool-button {
    border-color: #009FDF;
    background-color: #009FDF;
  }
  #user-tool-button:hover {
    border-color: rgb(0, 139, 194);
    background-color: rgb(0, 139, 194);
  }
</style>

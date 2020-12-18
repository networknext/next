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
      <div class="mb-2 mb-md-0 flex-grow-1 align-items-center pl-4 pr-4" v-if="$store.getters.isAnonymousPlus">
        <Alert :message="`Please confirm your email address: ${$store.getters.userProfile.email}`" :alertType="AlertType.INFO" ref="verifyAlert">
          <a href="#" @click="$refs.verifyAlert.resendVerificationEmail()">
            Resend email
          </a>
        </Alert>
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
          <div class="col-auto">
            <button class="btn btn-primary" type="submit">
              View Stats
            </button>
          </div>
        </div>
      </div>
    </form>
    <Alert :message="message" :alertType="alertType" v-if="message !== '' && $route.path === '/session-tool'"/>
    <router-view :key="$route.fullPath"/>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Route, NavigationGuardNext } from 'vue-router'

import Alert from '@/components/Alert.vue'
import { AlertType } from '@/components/types/AlertTypes'
/**
 * This component holds the workspace elements related to the session tool in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert
  }
})
export default class SessionToolWorkspace extends Vue {
  private alertType: string
  private message: string
  private searchID: string
  private AlertType: any

  constructor () {
    super()
    this.alertType = ''
    this.searchID = ''
    this.message = 'Please enter a valid Session ID to view its statistics. It should be a hexadecimal number (with leading zeros), or a decimal number.'
    this.alertType = AlertType.INFO
    this.AlertType = AlertType
  }

  private created () {
    console.log('Created')
    this.searchID = this.$route.params.pathMatch || ''
  }

  private mounted () {
    console.log('mounted')
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    this.searchID = to.params.pathMatch || ''
    this.message = 'Please enter a valid Session ID to view its statistics. It should be a hexadecimal number (with leading zeros), or a decimal number.'
    this.alertType = AlertType.INFO
    next()
  }

  private fetchSessionDetails () {
    this.message = ''
    if (this.searchID === '') {
      this.$router.push({ path: '/session-tool' })
      return
    }
    const newRoute = `/session-tool/${this.searchID}`
    if (this.$route.path !== newRoute) {
      this.$router.push({ path: newRoute })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

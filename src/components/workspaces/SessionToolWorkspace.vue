<template>
  <main role="main" class="col-md-12 col-lg-12 px-4">
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
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
        <div class="mr-auto"></div>
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
    <router-view />
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Route, NavigationGuardNext } from 'vue-router'

/**
 * TODO: Cleanup template
 * TODO: Figure out what sessionMeta fields need to be required
 */

@Component
export default class SessionToolWorkspace extends Vue {
  // TODO: Refactor out the alert/error into its own component.

  private searchID: string
  private message: string
  private alertType: string

  private constructor () {
    super()
    this.message = ''
    this.alertType = ''
    this.searchID = ''
  }

  private created () {
    // Empty for now
    this.searchID = this.$route.params.pathMatch || ''
    this.message = this.searchID === '' ? '' : 'Please enter a valid Session ID to view its statistics. It should be a hexadecimal number (with leading zeros), or a decimal number.'
    this.alertType = this.searchID === '' ? '' : 'alert-info'
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    this.searchID = ''
    next()
  }

  private fetchSessionDetails () {
    this.message = ''
    this.alertType = ''
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

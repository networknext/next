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
                   v-model="searchInput"
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
    <!-- TODO: Refactor these into a seperate component -->
    <div class="alert alert-info" role="alert" id="session-tool-alert" v-if="showAlert">
      Please enter a valid Session ID to view its statistics.
      It should be a hexadecimal number (with leading zeros), or a decimal number.
    </div>
    <div class="alert alert-danger" role="alert" id="session-tool-danger" v-if="showError">
      Failed to fetch session details
    </div>
    <SessionDetails v-if="searchID != ''" v-bind:searchID="searchID"/>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Route, NavigationGuardNext } from 'vue-router'
import SessionDetails from '@/components/SessionDetails.vue'

/**
 * TODO: Cleanup template
 * TODO: Figure out what sessionMeta fields need to be required
 */

@Component({
  components: {
    SessionDetails
  }
})
export default class SessionToolWorkspace extends Vue {
  // TODO: Refactor out the alert/error into its own component.
  // TODO: Figure out how to make searchID and searchInput the same

  private showAlert = false
  private showError = false

  private searchID = ''
  private searchInput = ''
  private showDetails = false

  private created () {
    // Empty for now
    this.searchID = this.$route.params.id || ''
    this.searchInput = this.searchID
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    if (!to.params.id) {
      this.searchInput = ''
      this.searchID = ''
    }
    next()
  }

  private fetchSessionDetails () {
    this.searchID = this.searchInput
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

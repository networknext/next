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
    <form class="flow-stats-form" @submit.prevent="loadSessionDetails()">
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
    <router-view/>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * TODO: Cleanup template
 * TODO: Figure out what sessionMeta fields need to be required
 */

@Component
export default class SessionToolWorkspace extends Vue {
  // TODO: Refactor out the alert/error into its own component.
  private showAlert = false
  private showError = false

  private searchID = ''
  private showDetails = false

  private created () {
    // Empty for now
  }

  private loadSessionDetails () {
    // API Call to fetch the details associated to ID
    this.$router.push(`/session-tool/${this.searchID}`)
  }

  private beforeRouteUpdate (to: any, from: any, next: any) {
    if (to.params.pathMatch) {
      this.searchID = to.params.pathMatch
    } else {
      this.searchID = ''
    }
    next()
  }
}

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

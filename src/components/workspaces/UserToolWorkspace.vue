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
        User Tool
      </h1>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
        <div class="mr-auto"></div>
      </div>
    </div>
    <form class="flow-stats-form" @submit.prevent="fetchUserSessions()">
      <div class="form-group">
        <label for="user-hash-input">
            User Hash
        </label>
        <div class="row">
          <div class="col">
            <input class="form-control"
                    type="text"
                    placeholder="Enter a User Hash to view statistics"
                    v-model="searchID"
            >
          </div>
          <div class="col-auto">
            <button class="btn btn-primary" type="submit">
              View Sessions
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
import UserSessions from '@/components/UserSessions.vue'

@Component({
  components: {
    UserSessions
  }
})
export default class UserToolWorkspace extends Vue {
  // TODO: Refactor out the alert/error into its own component.
  private showAlert = false
  private showError = false

  private searchID = ''
  private showDetails = false

  private message = ''
  private alertType = ''

  private created () {
    // Empty for now
    this.searchID = this.$route.params.pathMatch || ''
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    this.searchID = ''
    next()
  }

  private fetchUserSessions () {
    this.message = ''
    this.alertType = ''
    if (this.searchID === '') {
      this.$router.push({ path: '/user-tool' })
      return
    }
    const newRoute = `/user-tool/${this.searchID}`
    if (this.$route.path !== newRoute) {
      this.$router.push({ path: newRoute })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

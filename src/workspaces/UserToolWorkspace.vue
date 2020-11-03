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
                    placeholder="Enter a User Id to view their sessions"
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
    <Alert :message="message" :alertType="alertType" v-if="alertMessage !== '' && $route.path === '/user-tool'"/>
    <router-view />
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import UserSessions from '@/components/UserSessions.vue'
import { AlertTypes } from '@/components/types/AlertTypes'
import Alert from '@/components/Alert.vue'

/**
 * This component holds the workspace elements related to the user tool page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component({
  components: {
    Alert,
    UserSessions
  }
})
export default class UserToolWorkspace extends Vue {
  get alertMessage () {
    return this.message
  }

  private alertType: string
  private message: string
  private searchID: string

  constructor () {
    super()
    this.alertType = ''
    this.searchID = ''
    this.message = 'Please enter a User Id to view their sessions.'
    this.alertType = AlertTypes.INFO
  }

  private created () {
    this.searchID = this.$route.params.pathMatch || ''
  }

  private fetchUserSessions () {
    this.message = ''
    if (this.searchID === '' && this.$route.path !== '/user-tool') {
      this.$router.push({ path: '/user-tool' })
      return
    }
    if (this.searchID === '' && this.$route.path === '/user-tool') {
      this.message = 'Please enter a User Id to view their sessions.'
      this.alertType = AlertTypes.INFO
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

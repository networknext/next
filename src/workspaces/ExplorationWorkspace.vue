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
        Data Exploration
      </h1>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
        <div class="mr-auto"></div>
      </div>
    </div>
    <div class="card" style="margin-bottom: 250px;">
      <div class="card-header">
        <ul class="nav nav-tabs card-header-tabs">
          <li class="nav-item">
            <router-link to="/explore/notifications" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'notifications'}">Notifications</router-link>
          </li>
          <li class="nav-item">
            <router-link to="/explore/analytics" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'analytics'}">Analytics</router-link>
          </li>
          <li class="nav-item">
            <router-link to="/explore/invoicing" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'invoicing'}">Invoicing</router-link>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
    <!-- <div class="card mb-2">
      <div class="card-body">
        <iframe
          src="https://networknextexternal.cloud.looker.com/login/embed/%2Fembed%2Flooks%2F1?access_filters=%7B%7D&external_user_id=%22auth0%7C5e823e827e97a90cf402109e%22&first_name=&force_logout_login=true&last_name=&models=%5B%22billing%22%5D&nonce=%22erC4QrcB3A9QAhUX%22&permissions=%5B%22access_data%22%2C%22see_looks%22%2C%22explore%22%2C%22see_drill_overlay%22%5D&session_length=3600&signature=nu4KKUi2Slf8lPU%2FYUnTA2I7RBs%3D&time=1624548915"
          width="600"
          height="300"
          frameborder="0">
        </iframe>
      </div>
    </div> -->
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

/**
 * This component holds the workspace elements related to the downloads page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component
export default class ExplorationWorkspace extends Vue {}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

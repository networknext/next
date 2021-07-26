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
        Explore
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
          <li class="nav-item" v-if="false">
            <router-link to="/explore/analytics" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'analytics'}">Analytics</router-link>
          </li>
          <li class="nav-item" v-if="false">
            <router-link to="/explore/invoicing" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'invoicing'}">Invoicing</router-link>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
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

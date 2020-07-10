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
    ">
      <h1 class="h2">
        Settings
      </h1>
      <div class="btn-toolbar mb-2 mb-md-0 flex-grow-1 hidden">
        <div class="mr-auto"></div>
      </div>
    </div>
    <div class="card" style="margin-bottom: 250px;">
      <div class="card-header">
        <ul class="nav nav-tabs card-header-tabs">
          <li class="nav-item">
            <router-link to="/settings/users" class="nav-link" v-bind:class="{ active: currentPage == 'users'}">Users</router-link>
          </li>
          <li class="nav-item">
            <router-link to="/settings/game-config" class="nav-link" v-bind:class="{ active: currentPage == 'game-config'}">Game Configuration</router-link>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Route, NavigationGuardNext } from 'vue-router'

@Component
export default class SettingsWorkspace extends Vue {
  private currentPage = ''

  private beforeRouteEnter (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    if (to.name === 'Settings') {
      next('/settings/users')
    }
    next()
  }

  private beforeRouteUpdate (to: Route, from: Route, next: NavigationGuardNext<Vue>) {
    const toName = to.name || ''
    const fromName = from.name || ''

    if (toName === 'Settings' && (fromName !== 'User Management' || fromName !== 'Game Configuration')) {
      // TODO: Check perms are route user to correct default tab
      next('/settings/users')
    }
    if (toName === 'Settings' && (fromName === 'User Management' || fromName === 'Game Configuration')) {
      // If the user clicks the settings tab when already in settings do nothing
      return
    }
    next()
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

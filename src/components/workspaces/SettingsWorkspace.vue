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
            <a class="nav-link" id="accounts-link" href="#" v-bind:class="{ active: currentPage == 'users' }">
              Users
            </a>
          </li>
          <li class="nav-item">
            <a class="nav-link" id="config-link" href="#" v-bind:class="{ active: currentPage == 'game-config' }">
              Game Configuration
            </a>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'

@Component
export default class SettingsWorkspace extends Vue {
  private currentPage = ''

  private beforeRouteUpdate (to: any, from: any, next: any) {
    // Redirect to /settings/users for now - this will be determined by permissions
    console.log('before route update')
    console.log(typeof to)
    console.log(to)
    console.log(typeof from)
    console.log(from)
    next()
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

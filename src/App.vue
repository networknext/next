<template>
  <div id="app">
    <nav-bar></nav-bar>
    <div style="padding-top: 48px;">
      <div class="alert alert-primary"
            style="text-align: center;"
            role="alert"
            v-if="false"
      >
        Please confirm your email address: EMAIL ADDRESS
        <a href="#">
            Resend email
        </a>
      </div>
      <div class="alert alert-success"
          style="text-align: center;"
          role="alert"
          v-if="false"
      >
        Verification email was sent successfully. Please check your email for futher instructions.
      </div>
      <div class="alert alert-danger"
          style="text-align: center;"
          role="alert"
          v-if="false"
      >
        Something went wrong sending the verification email. Please try again later.
      </div>
    </div>
    <div class="container-fluid">
      <div class="row">
        <map-workspace></map-workspace>
        <sessions-workspace v-if="false"></sessions-workspace>
        <session-tool-workspace v-if="false"></session-tool-workspace>
        <user-tool-workspace v-if="false"></user-tool-workspace>
        <downloads-workspace v-if="false"></downloads-workspace>
        <settings-workspace v-if="false"></settings-workspace>
      </div>
    </div>
  </div>
</template>

<style lang="scss">
  @import url('https://fonts.googleapis.com/css?family=Montserrat:400,700');
  .fill {
    width: 100%;
    height: 100%;
  }
  .dashboard-title {
    width: 95%;
  }
  .workspace {
    width: 100%;
    justify-content: center;
  }
  .map {
    height: 95%;
    max-height: 1000px;
  }
  .mapboxgl-canvas {
    height: 100%;
    width: 100%;
  }
  .mapboxgl-map {
    height: 100%;
    width: 100%;
  }
  #session-tool-map { position: absolute; top: 0; bottom: 0; width: 100%; }
  body {
    font-size: 0.875rem;
  }

  .feather {
    width: 16px;
    height: 16px;
    vertical-align: text-bottom;
  }
  .logo-container {
    width: 100%;
    height: 24px;
    margin: 0;
    padding: 0;
    display: flex;
    align-items: center;
    justify-content: left;
  }
  .logo-fit {
    max-height: 24px;
    max-width: 100%;
  }
  .fixed-width {
    font-family: monospace;
    font-size: 120%;
  }
  .hidden {
    display: none;
  }
  div.table-no-top-line th {
    border-top: none !important;
  }
  .mapboxgl-ctrl-bottom-left {
    display: none;
  }
  .mapboxgl-ctrl-bottom-right {
    display: none;
  }
  .h2 {
    font-size: 2rem;
  }

  /*
  * Sidebar
  */

  .sidebar {
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    z-index: 100; /* Behind the navbar */
    padding: 48px 0 0; /* Height of navbar */
    box-shadow: inset -1px 0 0 rgba(0, 0, 0, 0.1);
  }

  .sidebar-sticky {
    position: relative;
    top: 0;
    height: calc(100vh - 48px);
    padding-top: 0.5rem;
    overflow-x: hidden;
    overflow-y: auto; /* Scrollable contents if viewport is shorter than content. */
  }

  @supports ((position: -webkit-sticky) or (position: sticky)) {
    .sidebar-sticky {
      position: -webkit-sticky;
      position: sticky;
    }
  }

  .sidebar .nav-link {
    font-weight: 500;
    color: #333;
  }

  .sidebar .nav-link .feather {
    margin-right: 4px;
    color: #999;
  }

  .sidebar .nav-link.active {
    color: #007bff;
  }

  .sidebar .nav-link:hover .feather,
  .sidebar .nav-link.active .feather {
    color: inherit;
  }

  .sidebar-heading {
    font-size: 0.75rem;
    text-transform: uppercase;
  }
</style>
<script lang="ts">
import Vue from 'vue'

import {
  faCheck,
  faCircle,
  faDownload,
  faPen,
  faTimes,
  faTrash
} from '@fortawesome/free-solid-svg-icons'
import { library } from '@fortawesome/fontawesome-svg-core'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import NavBar from './components/NavBar.vue'
import MapWorkspace from './components/MapWorkspace.vue'
import SessionsWorkspace from './components/SessionsWorkspace.vue'
import SessionToolWorkspace from './components/SessionToolWorkspace.vue'
import UserToolWorkspace from './components/UserToolWorkspace.vue'
import DownloadsWorkspace from './components/DownloadsWorkspace.vue'
import SettingsWorkspace from './components/SettingsWorkspace.vue'

const ICONS = [
  faCheck,
  faCircle,
  faDownload,
  faPen,
  faTimes,
  faTrash
]

library.add(...ICONS)

Vue.component('font-awesome-icon', FontAwesomeIcon)
Vue.component('nav-bar', NavBar)
Vue.component('map-workspace', MapWorkspace)
Vue.component('sessions-workspace', SessionsWorkspace)
Vue.component('session-tool-workspace', SessionToolWorkspace)
Vue.component('user-tool-workspace', UserToolWorkspace)
Vue.component('downloads-workspace', DownloadsWorkspace)
Vue.component('settings-workspace', SettingsWorkspace)
export default Vue.extend({
  beforeCreate: () => {
    Vue.prototype.$apiService.call('BuyersService.TopSessions', {}).then((response: any) => {
      console.log(response)
    })
  }
})

</script>

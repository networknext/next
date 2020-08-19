<template>
  <div id="app">
    <nav-bar></nav-bar>
    <workspace></workspace>
  </div>
</template>

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
import Workspace from './components/workspaces/Workspace.vue'

/**
 * Main component that is used to mount the entire Vue app. Contains the
 *  main layout components (navbar and workspace)
 * This component is also responsible for mounting third party libs as well
 *  and defining global styles
*/

/**
 * TODO: Purge unnecessary styles
 */

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
Vue.component('workspace', Workspace)
export default Vue.extend({})

</script>

<style lang="scss">
  @import url('https://fonts.googleapis.com/css?family=Montserrat:400,700');
  html, body {
    height: 100%;
    overflow: hidden;
  }
  #app {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    overflow: auto;
  }
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

  .below-nav-bar {
    padding-top: 48px;
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

  .nav-link {
    font-weight: 500;
    color: #333;
  }

  .nav-link .feather {
    margin-right: 4px;
    color: #999;
  }

  .nav-link.active {
    color: #007bff;
  }

  .nav-link:hover .feather,
  .nav-link.active .feather {
    color: inherit;
  }

  .sidebar-heading {
    font-size: 0.75rem;
    text-transform: uppercase;
  }
</style>

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

  .map-container-offset {
    width: 100%;
    height: calc(-160px + 90vh);
    border: 1px solid rgb(136, 136, 136);
    background-color: rgb(27, 27, 27);
  }
  .map-container-no-offset {
    width: 100%;
    height: calc(-160px + 100vh);
    border: 1px solid rgb(136, 136, 136);
    background-color: rgb(27, 27, 27);
  }
  .uplot .legend.inline th::after {
      content: "";
      vertical-align: middle;
  }
  .choices__list--multiple .choices__item {
      background-color: lightgray;
      border: 1px solid darkgray;
      color: black;
      border-radius: 5px;
  }
  .choices[data-type*="select-multiple"]
  .choices__button, .choices[data-type*="text"] .choices__button {
      border-left: 0px;
      color: black;
  }
  .choices__item .is-highlighted {
      background-color: lightgray;
      border: 1px solid darkgray;
      color: black;
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

  .form-control-dark {
      color: #fff;
      background-color: rgba(255, 255, 255, 0.1);
      border-color: rgba(255, 255, 255, 0.1);
  }

  .form-control-dark:focus {
      border-color: transparent;
      box-shadow: 0 0 0 3px rgba(255, 255, 255, 0.25);
  }

  .btn-group-xs > .btn,
  .btn-xs {
      padding: 0.25rem 0.4rem;
      font-size: 0.875rem;
      line-height: 0.5;
      border-radius: 0.2rem;
  }

  td.td-btn {
      padding-top: 0.15rem;
      padding-bottom: 0;
  }

  td.td-btn > a.btn-xs {
      padding-bottom: calc(0.25rem + 1px);
  }
</style>
<script lang="ts">
import Vue from 'vue';

import {
  faCheck,
  faCircle,
  faDownload,
  faPen,
  faTimes,
  faTrash,
} from '@fortawesome/free-solid-svg-icons';
import { library } from '@fortawesome/fontawesome-svg-core';
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome';
import NavBar from './components/NavBar.vue';
import MapWorkspace from './components/MapWorkspace.vue';
import SessionsWorkspace from './components/SessionsWorkspace.vue';
import SessionToolWorkspace from './components/SessionToolWorkspace.vue';
import UserToolWorkspace from './components/UserToolWorkspace.vue';
import DownloadsWorkspace from './components/DownloadsWorkspace.vue';
import SettingsWorkspace from './components/SettingsWorkspace.vue';

const ICONS = [
  faCheck,
  faCircle,
  faDownload,
  faPen,
  faTimes,
  faTrash,
];

library.add(...ICONS);

Vue.component('font-awesome-icon', FontAwesomeIcon);
Vue.component('nav-bar', NavBar);
Vue.component('map-workspace', MapWorkspace);
Vue.component('sessions-workspace', SessionsWorkspace);
Vue.component('session-tool-workspace', SessionToolWorkspace);
Vue.component('user-tool-workspace', UserToolWorkspace);
Vue.component('downloads-workspace', DownloadsWorkspace);
Vue.component('settings-workspace', SettingsWorkspace);
export default Vue.extend({
  beforeCreate: () => {
    Vue.prototype.$apiService.call('BuyersService.TopSessions', {}).then((response: any) => {
      console.log(response);
    });
  },
});

</script>

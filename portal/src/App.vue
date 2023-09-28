// --------------------------------------------------------------------

<template>

  <nav class="navbar fixed-top navbar-expand-xxl navbar-light bg-light">

    <div class="container-fluid">

      <div style="height=60px">
        <div class="d-xxl-none">
          <router-link class="navbar-brand" to="/"><img src="@/assets/n_black.png" style="width: 26.4; height:30px;"/></router-link>
          <SessionCounts/>
        </div>
        <div class="d-none d-xxl-block">
          <router-link class="navbar-brand" to="/"><img src="@/assets/logo_black.png" style="width: 256.526946075px; height:30px;"/></router-link>
        </div>
      </div>

      <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarSupportedContent" @click="visible=!visible">
        <span class="navbar-toggler-icon"></span>
      </button>

      <div class="navbar-collapse" :class="!visible?'collapse':''" id="navbarSupportedContent">

        <ul class="navbar-nav me-auto mb-2 mb-lg-0">

          <li class="nav-item">
            <router-link class="nav-link" to="/" @click="visible=!visible">Map</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" to="/sessions" @click="visible=!visible">Sessions</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" to="/servers" @click="visible=!visible">Servers</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" to="/relays" @click="visible=!visible">Relays</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" to="/datacenters" @click="visible=!visible">Datacenters</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" to="/buyers" @click="visible=!visible">Buyers</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" to="/sellers" @click="visible=!visible">Sellers</router-link>
          </li>

        </ul>

        <div class="d-none d-xxl-block">
          <SessionCounts/>
        </div>

      </div>
    </div>
  </nav>

  <router-view/>

</template>

// --------------------------------------------------------------------

<script>

import SessionCounts from '@/components/SessionCounts.vue';

export default {
  name: 'App',
  components: {
    SessionCounts,
  },
  data() {
    return {
      visible: false,
    }
  },
}
</script>

// --------------------------------------------------------------------

<style lang="scss">

  @import url('https://fonts.googleapis.com/css?family=Montserrat:400,700');

  .navbar a{
    font-family: "Montserrat" !important;
    font-weight: 10;
    font-size: 18px;
    -webkit-user-drag: none;
  }

  html, body {
    height: 100%;
    overflow: hidden;
    font-family: "Montserrat" !important;
    font-weight: 400;
    font-style: normal;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
    -webkit-user-drag: none;
    font-size: 85%;
  }

  #app {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    overflow: auto;
    text-align: center;
    color: #2c3e50;
    margin-top: 0px;
    padding-top: 75px;
    width: 100%;
  }

  .table-hover tbody tr:hover td, .table-hover tbody tr:hover th {
    background-color: #439EFF33;
  }

</style>

// --------------------------------------------------------------------

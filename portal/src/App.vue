// --------------------------------------------------------------------

<template>

  <nav class="navbar fixed-top navbar-expand-xxl navbar-light bg-light">

    <div class="container-fluid">

      <div>
        <div class="d-xxl-none">
          <router-link class="navbar-brand" @click="nav_map()" to="/map"><img src="@/assets/n_black.png" style="width: 26.4; height:30px;"/></router-link>
          <SessionCounts/>
        </div>
        <div class="d-none d-xxl-block">
          <router-link class="navbar-brand" @click="nav_map()" to="/map"><img src="@/assets/logo_black.png" style="width: 256.526946075px; height:30px;"/></router-link>
        </div>
      </div>

      <div v-if="num_pages > 1" class="d-xxl-none arrows">
        <i class="bi bi-arrow-left-circle material-icons" @click="prevPage()"></i>
        <i class="bi bi-arrow-right-circle material-icons" @click="nextPage()"></i>
      </div>

      <button class="navbar-toggler" type="button" @click="this.visible=!this.visible">
        <span class="navbar-toggler-icon"></span>
      </button>

      <div class="navbar-collapse" :class="!this.visible?'collapse':''" id="navbarSupportedContent">

        <ul class="navbar-nav me-auto mb-2 mb-lg-0">

          <li class="nav-item">
            <router-link class="nav-link" @click="nav_map()" to="/map">Map</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" @click="nav_sessions()" to="/sessions">Sessions</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" @click="nav_servers()" to="/servers">Servers</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" @click="nav_relays()" to="/relays">Relays</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" @click="nav_datacenters" to="/datacenters">Datacenters</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" @click="nav_buyers" to="/buyers">Buyers</router-link>
          </li>

          <li class="nav-item">
            <router-link class="nav-link" @click="nav_sellers" to="/sellers">Sellers</router-link>
          </li>

        </ul>

        <div class="d-none d-xxl-block">
          <SessionCounts/>
        </div>

      </div>
    </div>
  </nav>

  <router-view id='router' @notify-update="onUpdate" @notify-loaded="onLoaded" @notify-view="onView"/>

</template>

// --------------------------------------------------------------------

<script>

import SessionCounts from '@/components/SessionCounts.vue';
import emitter from "@/mitt.js";

export default {

  name: 'App',

  components: {
    SessionCounts,
  },

  data() {
    return {
      visible: false,
      page: 0,
      num_pages: 1,
      current_view: 'map',
    }
  },

  methods: {

    nav_map() {
      if (this.current_view != 'map') {
        this.$router.push('/map')
      } else {
        this.onLoaded()
      }
    },

    nav_sessions() {
      if (this.current_view != 'sessions') {
        this.$router.push('/sessions')
      } else {
        this.onLoaded()
      }
    },

    nav_servers() {
      if (this.current_view != 'servers') {
        this.$router.push('/servers')
      } else {
        this.onLoaded()
      }
    },

    nav_relays() {
      if (this.current_view != 'relays') {
        this.$router.push('/relays')
      } else {
        this.onLoaded()
      }
    },

    nav_datacenters() {
      if (this.current_view != 'datacenters') {
        this.$router.push('/datacenters')
      } else {
        this.onLoaded()
      }
    },

    nav_buyers() {
      if (this.current_view != 'buyers') {
        this.$router.push('/buyers')
      } else {
        this.onLoaded()
      }
    },

    nav_sellers() {
      if (this.current_view != 'sellers') {
        this.$router.push('/sellers')
      } else {
        this.onLoaded()
      }
    },

    onUpdate(page, num_pages) {
      this.page = page
      this.num_pages = num_pages
    },

    onLoaded() {
      this.visible = false
    },

    onView(current) {
      this.current_view = current
    },

    prevPage() {
      emitter.emit('prev_page')
    },

    nextPage() {
      emitter.emit('next_page')
    },

  }
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
  padding-top: 45px;
  width: 100%;
}

.arrows {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  height: 10px;
}

.table-hover tbody tr:hover td, .table-hover tbody tr:hover th {
  background-color: #439EFF33;
}

.material-icons {
    color: #bbbbbbcc;
    font-size: 38px;
}

.left-align {
  text-align: left;
}

.right_align {
  text-align: right;
}

</style>

// --------------------------------------------------------------------

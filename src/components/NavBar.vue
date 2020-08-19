<template>
  <nav class="navbar navbar-expand-md navbar-dark fixed-top bg-dark p-0 shadow">
    <a class="navbar-brand col-sm-3 col-md-2 mr-0" href="https://networknext.com">
      <div class="logo-container">
        <img class="logo-fit" src="../assets/logo.png">
      </div>
    </a>
    <ul class="navbar-nav px-3 w-100 mr-auto">
      <li class="nav-item text-nowrap">
        <router-link to="/" class="nav-link" v-bind:class="{ active: $store.getters.currentPage == 'map' }" data-test="mapLink">Map</router-link>
      </li>
      <li class="nav-item text-nowrap">
        <router-link to="/sessions" class="nav-link" v-bind:class="{ active: $store.getters.currentPage == 'sessions' }" data-test="sessionsLink">Sessions</router-link>
      </li>
      <li class="nav-item text-nowrap">
        <router-link to="/session-tool"
                      class="nav-link"
                      v-bind:class="{
                        active: $store.getters.currentPage == 'session-tool' || $store.getters.currentPage == 'session-details'
                      }"
                      data-test="sessionToolLink">
                      Session Tool
        </router-link>
      </li>
      <li class="nav-item text-nowrap">
        <router-link to="/user-tool"
                     class="nav-link"
                     v-bind:class="{
                        active: $store.getters.currentPage == 'user-tool' || $store.getters.currentPage == 'user-sessions'
                     }"
                     v-if="!$store.getters.isAnonymous">
                      User Tool
        </router-link>
      </li>
      <li class="nav-item text-nowrap">
        <router-link to="/downloads" class="nav-link" v-bind:class="{ active: $store.getters.currentPage == 'downloads' }" v-if="!$store.getters.isAnonymous">Downloads</router-link>
      </li>
      <li class="nav-item text-nowrap">
        <router-link to="/settings" class="nav-link" v-bind:class="{ active: $store.getters.currentPage == 'config' || $store.getters.currentPage == 'users' }" v-if="!$store.getters.isAnonymous">Settings</router-link>
      </li>
    </ul>
    <ul class="navbar-nav px-3 w-100">
      <li class="nav-item text-nowrap" style="color: #9a9da0;">
        {{ portalVersion }}
      </li>
    </ul>
    <ul class="navbar-nav px-3" v-if="$store.getters.isAnonymous">
      <li class="nav-item text-nowrap">
        <a data-test="loginButton" class="login btn-sm btn-primary" href="#" @click="authService.logIn()">
          Log in
        </a>
      </li>
    </ul>
    <ul class="navbar-nav px-3" v-if="$store.getters.isAnonymous">
      <li class="nav-item text-nowrap">
        <a data-test="signUpButton" class="signup btn-sm btn-primary" href="#" @click="authService.signUp()">
          Sign up
        </a>
      </li>
    </ul>
    <ul class="navbar-nav px-3" v-if="!$store.getters.isAnonymous">
      <li class="nav-item text-nowrap">
        <a class="logout btn-sm btn-primary" href="#" @click="authService.logOut()">
          Logout
        </a>
      </li>
    </ul>
  </nav>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import AuthService from '../services/auth.service'
import APIService from '@/services/api.service'
@Component
export default class NavBar extends Vue {
  private apiService: APIService
  private authService: AuthService
  private portalVersion: string

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
    this.authService = Vue.prototype.$authService
    this.portalVersion = ''
    this.fetchPortalVersion()
  }

  private fetchPortalVersion (): void {
    const isDev = window.location.hostname === 'portal-dev.networknext.com'
    if (this.$store.getters.isAdmin || isDev) {
      fetch(`${process.env.VUE_APP_API_URL}/version`, {
        headers: {
          Accept: 'application/json',
          'Accept-Encoding': 'gzip',
          'Content-Type': 'application/json'
        },
        method: 'POST'
      }).then((response: any) => {
        response.json().then((json: any) => {
          if (json.error) {
            throw new Error(json.error)
          }
          this.portalVersion = `Git Hash: ${json.sha} - Commit: ${json.commit_message || 'none'}`
        })
      })
    }
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
  .navbar-brand {
    display: inline-block;
    padding-top: .3125rem;
    padding-bottom: .3125rem;
    margin-right: 1rem;
    font-size: 1.25rem;
    line-height: inherit;
    white-space: nowrap;
  }
  .navbar .form-control {
    padding: 0.75rem 1rem;
    border-width: 0;
    border-radius: 0;
  }
  .signup {
    width: 6rem;
    height: 1.7rem;
    display: block;
    text-align: center;
    border-radius: 9999px;
    background-color: #f94a21;
    font-weight: bold;
    line-height: 1.1rem;
  }
  .signup:hover {
    text-decoration: none;
    background-color: rgba(249, 73, 33, 0.9);
  }
  .signup:not(:disabled):not(.disabled):active {
    background-color: #007bff;
    text-decoration: none;
    border-color: #007bff;
    box-shadow: none;
  }
  .signup:focus {
    background-color: #007bff;
    text-decoration: none;
    border-color: #007bff;
    box-shadow: none;
  }
  .login {
    width: 6rem;
    height: 1.7rem;
    display: block;
    border-width: 1px;
    border-color: #f94a21;
    border-radius: 9999px;
    text-align: center;
    background-color: #343a40;
    border-style: solid;
    font-weight: bold;
    line-height: 1rem;
  }
  .login:hover {
    background-color: rgba(249, 73, 33, 0.1);
    text-decoration: none;
    border-color: #f94a21;
  }
  .login:not(:disabled):not(.disabled):active {
    background-color: #343a40;
    text-decoration: none;
    border-color: #007bff;
    box-shadow: none;
  }
  .login:focus {
    background-color: rgba(0, 123, 255, 0.1);
    text-decoration: none;
    border-color: #007bff;
    box-shadow: none;
  }
  .logout {
    width: 6rem;
    height: 1.7rem;
    display: block;
    border-width: 1px;
    border-color: #f94a21;
    border-radius: 9999px;
    text-align: center;
    background-color: #343a40;
    border-style: solid;
    font-weight: bold;
    line-height: 1rem;
  }
  .logout:hover {
    background-color: rgba(249, 73, 33, 0.1);
    text-decoration: none;
    border-color: #f94a21;
  }
  .logout:not(:disabled):not(.disabled):active {
    background-color: #343a40;
    text-decoration: none;
    border-color: #007bff;
    box-shadow: none;
  }
  .logout:focus {
    background-color: rgba(0, 123, 255, 0.1);
    text-decoration: none;
    border-color: #007bff;
    box-shadow: none;
  }
</style>

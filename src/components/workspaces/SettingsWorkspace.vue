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
          <li class="nav-item" v-if="$store.getters.isAdmin || $store.getters.isOwner">
            <router-link to="/settings/users" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'users'}">Users</router-link>
          </li>
          <li class="nav-item">
            <router-link to="/settings/game-config" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'config'}">Game Configuration</router-link>
          </li>
          <li class="nav-item" v-if="$store.getters.isABTester">
            <router-link to="/settings/route-shader" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'shader'}">Route Shader</router-link>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import APIService from '../../services/api.service'

@Component
export default class SettingsWorkspace extends Vue {
  private apiService: APIService

  constructor () {
    super()
    this.apiService = Vue.prototype.$apiService
  }

  private mounted () {
    if (this.$store.getters.isAnonymous || this.$store.getters.isAnonymousPlus) {
      return
    }
    const userProfile = JSON.parse(JSON.stringify(this.$store.getters.userProfile))
    this.apiService
      .fetchGameConfiguration({ domain: this.$store.getters.userProfile.domain })
      .then((response: any) => {
        console.log(response)
        userProfile.pubKey = response.game_config.public_key
        userProfile.company = response.game_config.company
        userProfile.routeShader = response.customer_route_shader
        this.$store.commit('UPDATE_USER_PROFILE', userProfile)
      })
      .catch((e) => {
        console.log('Something went wrong fetching public key')
        console.log(e)
        this.$store.commit('UPDATE_USER_PROFILE', userProfile)
        userProfile.userProfile.pubKey = ''
        userProfile.userProfile.company = ''
      })
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>

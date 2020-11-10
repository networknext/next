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
          <li class="nav-item" v-if="!$store.getters.isAnonymous">
            <router-link to="/settings/account" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'account-settings'}">Account Settings</router-link>
          </li>
          <li class="nav-item" v-if="$store.getters.registeredToCompany && ($store.getters.isAdmin || $store.getters.isOwner)">
            <router-link to="/settings/users" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'users'}">Users</router-link>
          </li>
          <li class="nav-item" v-if="$store.getters.registeredToCompany && ($store.getters.isAdmin || $store.getters.isOwner)">
            <router-link to="/settings/game-config" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'config'}">Game Configuration</router-link>
          </li>
          <li class="nav-item" v-if="$flagService.isEnabled(featureEnums.ROUTE_SHADER) && ($store.getters.isAdmin || $store.getters.isOwner)">
            <router-link to="/settings/route-shader" class="nav-link" v-bind:class="{ active: $store.getters.currentPage === 'shader'}">Route Shader</router-link>
          </li>
        </ul>
      </div>
      <router-view/>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { FeatureEnum } from '../components/types/FeatureTypes'

/**
 * This component holds the workspace elements related to the settings page in the Portal
 */

/**
 * TODO: Make this a View
 */

@Component
export default class SettingsWorkspace extends Vue {
  // Empty for now
  private featureEnums: any

  constructor () {
    super()
    // TODO: This probably shouldn't be necessary?
    this.featureEnums = FeatureEnum
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped lang="scss">
</style>
